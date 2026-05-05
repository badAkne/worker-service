package config

import (
	"io"

	"github.com/badAkne/worker-service/internal/app/config/section"
	"github.com/badAkne/worker-service/internal/pkg/constant"
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type (
	// Config — главная структура конфигурации сервиса.
	Config struct {
		Repository section.Repository
		Processor  section.Processor
		Monitor    section.Monitor
		Meta       Meta `ignore:"true"`
		Broker     section.Broker
	}

	// Meta содержит метаданные о загрузке конфига.
	Meta struct {
		Load LoadArgs
	}

	// LoadArgs — аргументы для функции Load().
	LoadArgs struct {
		Output          io.Writer
		EnableSimpleLog bool
		SkipConfig      bool
	}
)

// Root — точка доступа к загруженным переменным конфигурации.
var Root Config

// Load загружает конфиг и инициализирует логгер сервиса.
func Load(args LoadArgs) {
	// Инициализируем логгер с debug уровнем, чтобы логировать ошибки загрузки конфига.
	zerolog.TimestampFieldName = "timestamp"
	zerolog.MessageFieldName = "msg"
	zerolog.TimeFieldFormat = constant.ISO8601

	if args.EnableSimpleLog {
		args.Output = zerolog.ConsoleWriter{Out: args.Output}
	}

	log.Logger = createLogger(zerolog.DebugLevel, args.Output)

	if args.SkipConfig {
		Root.Meta.Load = args
		return
	}

	// Загружаем .env файл.
	if err := godotenv.Load(); err != nil {
		log.Debug().Err(err).Msg("Файл .env не найден, используем переменные окружения")
	} else {
		log.Debug().Msg("Загружен .env файл")
	}

	// Парсим переменные окружения в структуру Config.
	if err := envconfig.Process(constant.EnvPrefix, &Root); err != nil {
		log.Fatal().Err(err).Msg("Не удалось загрузить конфиг из переменных окружения")
	}

	log.Debug().Msg("Конфиг загружен")

	// Пересоздаём логгер с настроенным уровнем логирования.
	logLevel, err := zerolog.ParseLevel(Root.Monitor.LogLevel)
	if err != nil {
		log.Warn().Err(err).Str("given_level", Root.Monitor.LogLevel).
			Msg("Некорректный уровень логирования, используем debug")
		logLevel = zerolog.DebugLevel
	}

	log.Logger = createLogger(logLevel, args.Output)

	// Сохраняем аргументы загрузки.
	Root.Meta.Load = args
}

// createLogger создаёт и возвращает zerolog.Logger с заданными параметрами.
func createLogger(level zerolog.Level, output io.Writer) zerolog.Logger {
	return zerolog.New(output).Level(level).With().Timestamp().Logger()
}
