package builder

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"reflect"
	"sync"
	"syscall"

	"github.com/badAkne/worker-service/internal/app/client/fixer"
	"github.com/badAkne/worker-service/internal/app/config"
	"github.com/badAkne/worker-service/internal/app/entity"
	ehandler "github.com/badAkne/worker-service/internal/app/handler/event"
	eorder "github.com/badAkne/worker-service/internal/app/handler/event/order"
	rhandler "github.com/badAkne/worker-service/internal/app/handler/http"
	"github.com/badAkne/worker-service/internal/app/handler/http/example"
	"github.com/badAkne/worker-service/internal/app/processor"
	eprocessor "github.com/badAkne/worker-service/internal/app/processor/event"
	rprocessor "github.com/badAkne/worker-service/internal/app/processor/http"
	pprocessor "github.com/badAkne/worker-service/internal/app/processor/other"
	"github.com/badAkne/worker-service/internal/app/repository"
	rcpostgres "github.com/badAkne/worker-service/internal/app/repository/conn/postgres"
	rcredis "github.com/badAkne/worker-service/internal/app/repository/conn/redis"
	rcurrency "github.com/badAkne/worker-service/internal/app/repository/currency"
	"github.com/badAkne/worker-service/internal/app/service"
	scurrency "github.com/badAkne/worker-service/internal/app/service/currency"
	sdelivery "github.com/badAkne/worker-service/internal/app/service/delivery"
	"github.com/badAkne/worker-service/internal/pkg/http/httph"
	"github.com/badAkne/worker-service/pkg/broker"
	"github.com/badAkne/worker-service/pkg/broker/codec"
	butil "github.com/badAkne/worker-service/pkg/broker/util"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
)

// Builder — структура для сборки зависимостей приложения.
// Использует паттерн Builder для последовательной инициализации компонентов.
type Builder struct {
	cCtx     *cli.Context
	ctx      context.Context
	wg       sync.WaitGroup
	err      error
	chErrors chan error

	// Подключения
	connPostgres *rcpostgres.Client
	connRedis    *rcredis.Client

	// Kafka
	brokerKafka                broker.KafkaClient
	busOrderCreated            broker.Bus[entity.EventOrderCreated]
	busOrderDeliveryCalculated broker.Bus[entity.EventOrderDeliveryCalculated]

	// Процессоры
	processors []processor.Processor

	// HTTP middleware (OpenTelemetry, NewRelic, и др.)
	middlewares []httph.Middleware

	// Handlers
	hExample          rhandler.Example
	handlerEventOrder ehandler.Order

	// repositories
	repoCurrencyRate repository.CurrencyRate
	//  services
	srvCurrency service.Currency
	srvDelivery service.Delivery
	//  clients
	fixerClient *fixer.Client

	// - handlers
	// - brokers (Kafka)
	// - monitors (OpenTelemetry, Prometheus)
}

// NewBuilder создаёт новый Builder и настраивает обработку сигналов OS.
// При получении SIGINT/SIGTERM контекст будет отменён.
func NewBuilder(cCtx *cli.Context) *Builder {
	b := Builder{cCtx: cCtx, chErrors: make(chan error, 4096)} // <- добавить chErrors
	var cancelFunc func()
	b.ctx, cancelFunc = context.WithCancel(context.Background())

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

	go b.waitForSignal(sig, cancelFunc)
	go b.printErrors() // <- добавить

	return &b
}

////////////////////////////////////////////////////////////////////////////////
///// CONFIG ///////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

// BuildConfig загружает конфигурацию из .env и переменных окружения.
// Можно передать injectors для модификации конфига после загрузки.
func (b *Builder) BuildConfig(injectors ...func(c *config.Config)) {
	b.buildConfig(config.LoadArgs{}, injectors)
}

// BuildConfigSimple загружает конфиг без файла .env (только injectors).
func (b *Builder) BuildConfigSimple(injectors ...func(c *config.Config)) {
	b.buildConfig(config.LoadArgs{SkipConfig: true}, injectors)
}

func (b *Builder) buildConfig(args config.LoadArgs, injectors []func(c *config.Config)) {
	if b.err != nil {
		return
	}

	// Определяем формат логов из CLI флага
	if b.cCtx != nil && b.cCtx.Bool("no-json") {
		args.EnableSimpleLog = true
	}
	args.Output = os.Stdout

	config.Load(args)

	// Применяем injectors
	for _, injector := range injectors {
		if injector != nil {
			injector(&config.Root)
		}
	}
}

////////////////////////////////////////////////////////////////////////////////
///// REPOSITORY CONNECTIONS ///////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

// BuildRepoConnPostgres инициализирует подключение к PostgreSQL.
func (b *Builder) BuildRepoConnPostgres() {
	b.exec(true, func(b *Builder) {
		cfg := config.Root.Repository.Postgres

		var err error
		b.connPostgres, err = rcpostgres.NewConn(b.ctx, cfg)
		if err != nil {
			b.err = fmt.Errorf("Repo.Conn.Postgres: %w", err)
			return
		}

		log.Debug().Msg("Unit Repo.Conn.Postgres has been initialized")
	})
}

// BuildRepoConnMigrator добавляет процессор миграций.
func (b *Builder) BuildRepoConnMigrator() {
	b.exec(b.connPostgres != nil, func(b *Builder) {
		proc := pprocessor.NewMigrator(b.connPostgres)
		b.processors = append(b.processors, proc)
	})
}

func (b *Builder) BuildConnRedis() {
	b.exec(true, func(b *Builder) {
		cfg := config.Root.Repository.Redis
		cl, err := rcredis.NewConn(b.ctx, cfg)
		if err != nil {
			b.err = fmt.Errorf("Repo.Conn.Redis: %w", err)
			return
		}

		b.connRedis = cl
	})
}

////////////////////////////////////////////////////////////////////////////////
///// BROKER AND BUSES /////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

func (b *Builder) BuildBrokerKafka() {
	b.exec(true, (*Builder).buildBrokerKafka)
}

func (b *Builder) buildBrokerKafka() {
	kafkaCfg := broker.KafkaConfig{
		Addresses:     config.Root.Broker.Kafka.Addresses,
		ConsumerGroup: config.Root.Broker.Kafka.ConsumerGroup,
		ClientID:      config.Root.Broker.Kafka.ClientID,
	}

	log.Debug().
		Any("addresses", kafkaCfg.Addresses).
		Str("group", kafkaCfg.ConsumerGroup).
		Msg("kafka config")

	kafkaClient, err := broker.NewKafkaClient(kafkaCfg)
	if err != nil {
		b.err = err
		return
	}

	b.brokerKafka = *kafkaClient
	type T1 = entity.EventOrderCreated
	type T2 = entity.EventOrderDeliveryCalculated

	codecOrderCreated := codec.NewCodecJson[T1]()
	codecOrderDeliveryCalculated := codec.NewCodecJson[T2]()

	b.busOrderCreated = broker.MustKafkaBus(&b.brokerKafka,
		codecOrderCreated,
		config.Root.Broker.Kafka.ModelOrder.Created.Topic,
		butil.Coalesce(config.Root.Broker.Kafka.ModelOrder.Created.ConsumerGroup,
			kafkaCfg.ConsumerGroup))

	b.busOrderDeliveryCalculated = broker.MustKafkaBus(
		&b.brokerKafka, codecOrderDeliveryCalculated,
		config.Root.Broker.Kafka.ModelOrder.DeliveryCalculated.Topic,
		"",
	)

	log.Debug().Msg("Kafka buses created")
}

////////////////////////////////////////////////////////////////////////////////
///// HANDLERS /////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

// BuildHandlerExample создаёт example handler.
func (b *Builder) BuildHandlerExample() {
	b.exec(true, func(b *Builder) {
		b.hExample = example.NewHandler()
		log.Debug().Msg("Unit Handler.Example has been initialized")
	})
}

func (b *Builder) BuildHandlerEventOrder() {
	b.exec(true, func(b *Builder) {
		b.handlerEventOrder = eorder.NewHandler(b.srvDelivery, b.busOrderDeliveryCalculated)
	}, b.srvDelivery, b.busOrderDeliveryCalculated)
}

// TODO: Добавить методы для других handlers:
// func (b *Builder) BuildHandlerUser() { ... }

////////////////////////////////////////////////////////////////////////////////
///// PROCESSORS ///////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

// BuildProcHttp создаёт и добавляет HTTP процессор.
func (b *Builder) BuildProcHttp() {
	b.exec(true, func(b *Builder) {
		cfg := config.Root.Processor.WebServer
		proc := rprocessor.NewHTTP(b.hExample, b.middlewares, cfg)
		b.processors = append(b.processors, proc)
	})
}

func (b *Builder) BuildProcEventSubscribeOrderCreated() {
	b.exec(true, (*Builder).buildProcEventSubscribeOrderCreated, b.handlerEventOrder, b.busOrderCreated)
}

func (b *Builder) buildProcEventSubscribeOrderCreated() {
	proc := eprocessor.NewOrderCreatedEventsCatcher(b.handlerEventOrder, b.busOrderCreated)
	b.processors = append(b.processors, proc)
	log.Info().Msg("Processor ORDER_CREATED registered")
}

////////////////////////////////////////////////////////////////////////////////
///// REPOSITORIES ///////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

func (b *Builder) BuildRepoCurrencyRate() {
	b.exec(true, (*Builder).buildRepoCurrencyRate, b.connRedis)
}

func (b *Builder) buildRepoCurrencyRate() {
	cfg := config.Root.Client.Fixer
	b.repoCurrencyRate = rcurrency.NewRedisRepository(b.connRedis.Cl, cfg.CacheTTL)
	fmt.Println("Currency rate repository created")
}

// //////////////////////////////////////////////////////////////////////////////
// /// SERVICES ///////////////////////////////////////////////////////////////
// //////////////////////////////////////////////////////////////////////////////
func (b *Builder) BuildServiceCurrency() {
	b.exec(true, (*Builder).buildModuleCurrency, b.fixerClient, b.repoCurrencyRate)
}

func (b *Builder) buildModuleCurrency() {
	b.srvCurrency = scurrency.NewService(b.fixerClient, b.repoCurrencyRate)
	fmt.Println("Currency service created")
}

func (b *Builder) BuildServiceDelivery() {
	b.exec(true, func(b *Builder) {
		b.srvDelivery = sdelivery.NewService(b.srvCurrency)
	}, b.srvCurrency)
}

// //////////////////////////////////////////////////////////////////////////////
// /// SERVICES ///////////////////////////////////////////////////////////////
// //////////////////////////////////////////////////////////////////////////////
func (b *Builder) BuildClientFixer() {
	b.exec(true, (*Builder).buildClientFixer)
}

func (b *Builder) buildClientFixer() {
	cfg := config.Root.Client.Fixer
	b.fixerClient = fixer.NewClient(cfg)
}

////////////////////////////////////////////////////////////////////////////////
///// RUN //////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

// Run запускает все подготовленные процессоры и ожидает их завершения.
func (b *Builder) Run() {
	if b.err != nil {
		log.Fatal().Err(b.err).Msg("Ошибка при инициализации приложения")
	}

	log.Info().Msg("Приложение инициализировано")
	defer log.Info().Msg("Приложение завершено, до свидания!")

	for _, proc := range b.processors {
		proc.StartAsync(b.ctx, &b.wg)
	}

	b.wg.Wait()
}

////////////////////////////////////////////////////////////////////////////////
///// INTERNAL /////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

// waitForSignal ожидает сигнал и вызывает cancelFunc.
func (b *Builder) waitForSignal(sig chan os.Signal, cancelFunc func()) {
	gotSig := <-sig
	log.Info().Str("sig", gotSig.String()).Msg("Запрошено завершение")
	cancelFunc()
}

// exec выполняет callback только если:
// - preCond == true
// - нет предыдущих ошибок
// - контекст не отменён
// - все requiredArgs не nil/zero
//

func (b *Builder) exec(preCond bool, cb func(b *Builder), requiredArgs ...any) {
	if !preCond || b.err != nil || b.ctx.Err() != nil {
		return
	}

	for _, requiredArg := range requiredArgs {
		rv := reflect.ValueOf(requiredArg)
		if rv.Type().Kind() == reflect.Struct || !rv.IsZero() {
			continue
		}

		b.err = fmt.Errorf("BUG: required %s, but empty", rv.Type().String())
		return
	}

	cb(b)
}

func (b *Builder) printErrors() {
	for err := range b.chErrors {
		log.Error().Err(err).Msg("Got new error")
	}
}
