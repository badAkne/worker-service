package main

import (
	"os"
	"strings"

	"github.com/urfave/cli/v2"

	"github.com/badAkne/worker-service/cmd"
)

const (
	// AppName — имя приложения.
	AppName = "worker-service"

	// AppVersion — версия приложения.
	AppVersion = "0.1.0"
)

func main() {
	const Usage = "MoM Boilerplate V2 — шаблон Go сервиса"

	const Description = `
Boilerplate для создания Go микросервисов.
Включает: HTTP сервер (gorilla/mux), Bun ORM, zerolog, миграции.

Примеры:
  ./app web-server    # Запустить HTTP сервер
  ./app migrate       # Применить миграции БД
`

	app := cli.App{
		Name:    AppName,
		Version: AppVersion,
		Usage:   Usage,
		Commands: []*cli.Command{
			cmd.WebServer(),
			cmd.Migrate(),
		},
		Description: strings.TrimSpace(Description),
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "no-json",
				Usage: "Человеко-читаемый формат для логов вместо JSON",
			},
		},
		EnableBashCompletion: true,
		CommandNotFound: func(cCtx *cli.Context, command string) {
			_ = cli.ShowAppHelp(cCtx)
		},
		HideHelpCommand: true,
	}

	if err := app.Run(os.Args); err != nil {
		_, _ = os.Stderr.WriteString(err.Error() + "\n")
		os.Exit(1)
	}
}
