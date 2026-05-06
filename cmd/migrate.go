package cmd

import (
	"strings"

	"github.com/badAkne/worker-service/internal/app/builder"
	"github.com/urfave/cli/v2"
)

const (
	cmdMigrateUsage = `Применяет миграции базы данных`

	cmdMigrateDescription = `
Устанавливает соединение с PostgreSQL базой данных и применяет
миграции, которые ещё не были применены.

ВНИМАНИЕ!
Не запускает веб-сервер, который мог бы предоставить
маршруты для проверки работоспособности (health check).
`
)

// Migrate возвращает CLI команду для выполнения миграций.
func Migrate() *cli.Command {
	return &cli.Command{
		Name:            "migrate",
		Aliases:         []string{"m"},
		Usage:           cmdMigrateUsage,
		Description:     strings.TrimSpace(cmdMigrateDescription),
		Action:          cmdMigrate,
		HideHelpCommand: true,
	}
}

// cmdMigrate — handler команды migrate.
func cmdMigrate(cCtx *cli.Context) error {
	app := builder.NewBuilder(cCtx)
	app.BuildConfig()

	app.BuildRepoConnPostgres()
	app.BuildRepoConnMigrator()

	app.Run()
	return nil
}
