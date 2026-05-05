package section

import "time"

type (
	// Repository — конфигурация источников данных.
	Repository struct {
		Postgres RepositoryPostgres
	}

	// RepositoryPostgres — конфигурация подключения к PostgreSQL.
	RepositoryPostgres struct {
		Address        string        `required:"true" default:"127.0.0.1:5432"`
		Username       string        `required:"true"`
		Password       string        `required:"true"`
		Name           string        `required:"true"`
		MigrationTable string        `split_words:"true" default:"schema_migrations"`
		ReadTimeout    time.Duration `split_words:"true" default:"30s"`
		WriteTimeout   time.Duration `split_words:"true" default:"30s"`
	}
)
