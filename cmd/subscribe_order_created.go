package cmd

import (
	"strings"

	"github.com/badAkne/worker-service/internal/app/builder"
	"github.com/urfave/cli/v2"
)

const (
	cmdSubscribeOrderCreatedUsage = "Подписка на топик событий создания заказов"

	cmdSubscribeOrderCreatedDescription = `
ВНИМАНИЕ!
Это команда подписки! Вы НЕ МОЖЕТЕ запустить эту команду в скрипт режиме.
Эта команда запускается только в режиме демона и работает до тех пор,
пока вы принудительно ее не остановите.

Команда подписывается на топик order.created для получения информации
о созданных заказах. Позволяет обрабатывать события и выполнять
дополнительные бизнес-операции (расчёт доставки, уведомления и т.д.).
`
)

func SubscribeOrderCreated() *cli.Command {
	return &cli.Command{
		Name:            "subscribe-order-created",
		Aliases:         []string{"suborder"},
		Usage:           cmdSubscribeOrderCreatedUsage,
		Description:     strings.TrimSpace(cmdSubscribeOrderCreatedDescription),
		Action:          cmdSubscribeOrderCreated,
		HideHelpCommand: true,
	}
}

func cmdSubscribeOrderCreated(cCtx *cli.Context) error {
	app := builder.NewBuilder(cCtx)
	app.BuildConfig()

	// Connections
	app.BuildConnRedis()
	app.BuildBrokerKafka()

	// Clients & Repositories
	app.BuildClientFixer()
	app.BuildRepoCurrencyRate()

	// Services
	app.BuildServiceCurrency()
	app.BuildServiceDelivery()

	// Handlers
	app.BuildHandlerEventOrder()

	// Processors
	app.BuildProcEventSubscribeOrderCreated()

	app.Run()
	return nil
}
