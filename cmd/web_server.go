package cmd

import (
	"strings"

	"github.com/badAkne/worker-service/internal/app/builder"
	"github.com/urfave/cli/v2"
)

const (
	cmdWebServerUsage = "Запуск HTTP сервера"

	cmdWebServerDescription = `
Запускает HTTP сервер для:
- Health check (/health)
- Метрики Prometheus (/metrics)
- Будущих API эндпоинтов
`
)

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

func cmdWebServer(cCtx *cli.Context) error {
	app := builder.NewBuilder(cCtx)
	app.BuildConfig()

	app.BuildConnRedis()

	// TODO: добавьте инициализацию компонентов для web-server
	// app.BuildProcHttp()

	app.Run()
	return nil
}
