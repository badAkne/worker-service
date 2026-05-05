package migration

import (
	"embed"
)

// Postgres содержит SQL миграции для PostgreSQL.
// Файлы должны находиться в папке postgres/ и называться:
//   - 001_description.up.sql   — применение миграции
//   - 001_description.down.sql — откат миграции
//
//go:embed postgres
var Postgres embed.FS
