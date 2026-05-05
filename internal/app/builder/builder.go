package builder

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"reflect"
	"sync"
	"syscall"

	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"

	"github.com/badAkne/worker-service/internal/app/config"
	rhandler "github.com/badAkne/worker-service/internal/app/handler/http"
	"github.com/badAkne/worker-service/internal/app/handler/http/example"
	"github.com/badAkne/worker-service/internal/app/processor"
	rprocessor "github.com/badAkne/worker-service/internal/app/processor/http"
	pprocessor "github.com/badAkne/worker-service/internal/app/processor/other"
	rcpostgres "github.com/badAkne/worker-service/internal/app/repository/conn/postgres"
	"github.com/badAkne/worker-service/internal/pkg/http/httph"
)

// Builder — структура для сборки зависимостей приложения.
// Использует паттерн Builder для последовательной инициализации компонентов.
type Builder struct {
	cCtx *cli.Context
	ctx  context.Context
	wg   sync.WaitGroup
	err  error

	// Подключения
	connPostgres *rcpostgres.Client

	// Процессоры
	processors []processor.Processor

	// HTTP middleware (OpenTelemetry, NewRelic, и др.)
	middlewares []httph.Middleware

	// Handlers
	hExample rhandler.Example

	// TODO: Добавить при необходимости:
	// - repositories
	// - modules
	// - handlers
	// - brokers (Kafka)
	// - monitors (OpenTelemetry, Prometheus)
}

// NewBuilder создаёт новый Builder и настраивает обработку сигналов OS.
// При получении SIGINT/SIGTERM контекст будет отменён.
func NewBuilder(cCtx *cli.Context) *Builder {
	b := Builder{cCtx: cCtx}
	var cancelFunc func()
	b.ctx, cancelFunc = context.WithCancel(context.Background())

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

	go b.waitForSignal(sig, cancelFunc)

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
//nolint:unparam // requiredArgs используется в других методах
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
