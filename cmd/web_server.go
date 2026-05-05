package cmd

import (
	"strings"

	"github.com/urfave/cli/v2"

	"github.com/badAkne/worker-service/internal/app/builder"
)

const (
	cmdWebServerUsage = "Запускает HTTP веб-сервер"

	cmdWebServerDescription = `
Инициализирует и запускает веб-сервер, который слушает указанный порт
для входящих HTTP запросов.
`
)

// WebServer возвращает CLI команду для запуска веб-сервера.
func WebServer() *cli.Command {
	return &cli.Command{
		Name:            "web-server",
		Aliases:         []string{"web", "http"},
		Usage:           cmdWebServerUsage,
		Description:     strings.TrimSpace(cmdWebServerDescription),
		Action:          cmdWebServer,
		HideHelpCommand: true,
	}
}

// cmdWebServer — handler команды web-server.
func cmdWebServer(cCtx *cli.Context) error {
	app := builder.NewBuilder(cCtx)
	app.BuildConfig()

	// Подключение к БД (опционально, раскомментируйте при необходимости)
	// app.BuildRepoConnPostgres()

	// TODO: Добавить репозитории и модули:
	// app.BuildRepoXxx()      // Репозитории
	// app.BuildModuleXxx()    // Модули бизнес-логики

	// Handlers
	app.BuildHandlerExample()

	// HTTP процессор
	app.BuildProcHttp()

	app.Run()
	return nil
}
