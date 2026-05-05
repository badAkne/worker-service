package rcpostgres

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/badAkne/worker-service/internal/app/config/section"
	"github.com/badAkne/worker-service/migration"
	"github.com/rs/zerolog/log"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"github.com/uptrace/bun/migrate"
)

type (
	// Client — обёртка над bun.DB с дополнительными методами.
	Client struct {
		_bunDB
		rawBunDB *bun.DB // Для служебных целей (миграции, транзакции)

		cfg section.RepositoryPostgres
	}

	_bunDB = bun.IDB
)

// GetRawBunDB возвращает оригинальный *bun.DB.
func (c *Client) GetRawBunDB() *bun.DB {
	return c.rawBunDB
}

// NewConn создаёт новое подключение к PostgreSQL.
func NewConn(ctx context.Context, cfg section.RepositoryPostgres) (*Client, error) {
	var u url.URL
	u.Scheme = "postgres"
	u.Host = cfg.Address
	u.User = url.UserPassword(cfg.Username, cfg.Password)
	u.Path = cfg.Name

	args := make(url.Values)
	args.Set("sslmode", "disable")

	u.RawQuery = args.Encode()

	log.Trace().
		Str("read_timeout", cfg.ReadTimeout.String()).
		Str("write_timeout", cfg.WriteTimeout.String()).
		Msg("Инициализация подключения к Postgres")

	sqlDB := sql.OpenDB(pgdriver.NewConnector(
		pgdriver.WithDSN(u.String()),
		pgdriver.WithReadTimeout(cfg.ReadTimeout),
		pgdriver.WithWriteTimeout(cfg.WriteTimeout),
	))
	sqlDB.SetMaxOpenConns(10)

	rawBunDB := bun.NewDB(sqlDB, pgdialect.New(), bun.WithDiscardUnknownColumns())

	var cancelFunc func()
	ctx, cancelFunc = context.WithTimeout(ctx, 2*time.Second)
	defer cancelFunc()

	if err := rawBunDB.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("не удалось проверить подключение: %w", err)
	}

	// TODO: Добавить bunotel.NewQueryHook для OpenTelemetry при необходимости
	// rawBunDB.AddQueryHook(bunotel.NewQueryHook(
	// 	bunotel.WithDBName(cfg.Name),
	// 	bunotel.WithFormattedQueries(true),
	// ))

	bunDB := newBunIdbTxInjector(rawBunDB)
	return &Client{cfg: cfg, _bunDB: bunDB, rawBunDB: rawBunDB}, nil
}

// Migrate выполняет миграции базы данных.
// Возвращает старую и новую версии схемы.
func (c *Client) Migrate(ctx context.Context) (oldVer, newVer int64, err error) {
	migrations := migrate.NewMigrations()

	if err = migrations.Discover(migration.Postgres); err != nil {
		return 0, 0, fmt.Errorf("не удалось обнаружить миграции: %w", err)
	}

	opts := []migrate.MigratorOption{
		migrate.WithTableName(c.cfg.MigrationTable),
		migrate.WithLocksTableName(c.cfg.MigrationTable + "_lock"),
		migrate.WithMarkAppliedOnSuccess(true),
	}

	m := migrate.NewMigrator(c.rawBunDB, migrations, opts...)

	if err = m.Init(ctx); err != nil {
		return 0, 0, fmt.Errorf("не удалось инициализировать migrator: %w", err)
	}

	// Получаем старую версию ДО миграции.
	applied, err := m.AppliedMigrations(ctx)
	if err != nil {
		return 0, 0, fmt.Errorf("не удалось получить применённые миграции: %w", err)
	}
	if len(applied) > 0 {
		// Список отсортирован по убыванию, берём первый элемент.
		oldVer, _ = strconv.ParseInt(applied[0].Name, 10, 64)
	}

	// Применяем миграции.
	mgg, err := m.Migrate(ctx)
	if err != nil {
		return 0, 0, fmt.Errorf("не удалось применить миграции: %w", err)
	}

	// Определяем новую версию миграции.
	newVer = oldVer
	for _, mg := range mgg.Migrations {
		ver, _ := strconv.ParseInt(mg.Name, 10, 64)
		if ver > newVer {
			newVer = ver
		}
	}

	return oldVer, newVer, nil
}

// InsideTx выполняет функцию внутри транзакции.
// Вложенные транзакции не поддерживаются — если транзакция уже есть в контексте,
// функция выполняется без создания новой транзакции.
func (c *Client) InsideTx(
	ctx context.Context, f func(ctx context.Context) error,
) error {
	// Не разрешаем вложенные транзакции.
	tx := getTxFromContext(ctx)
	if tx.Tx != nil {
		return f(ctx)
	}

	// Создаём транзакцию и помещаем её в контекст.
	var done bool
	var err error

	tx, err = c.rawBunDB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("не удалось начать транзакцию: %w", err)
	}

	// Откат транзакции в defer на случай паники.
	defer func() {
		if !done {
			_ = tx.Rollback()
		}
	}()

	ctx = setTxToContext(ctx, tx)

	// Выполняем функцию. Если ошибка — откатываем транзакцию.
	err = f(ctx)
	if err != nil {
		return err
	}

	// Устанавливаем done в true, чтобы ROLLBACK не был вызван.
	done = true

	// Коммитим транзакцию.
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("не удалось закоммитить транзакцию: %w", err)
	}

	return nil
}

// Close закрывает подключение к БД.
func (c *Client) Close() error {
	return c.rawBunDB.Close()
}
