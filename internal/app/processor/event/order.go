package eprocessor

import (
	"context"
	"sync"

	"github.com/badAkne/worker-service/internal/app/entity"
	ehandler "github.com/badAkne/worker-service/internal/app/handler/event"
	"github.com/badAkne/worker-service/internal/app/processor"
	"github.com/badAkne/worker-service/pkg/broker"
	"github.com/rs/zerolog/log"
)

type orderCreatedProc struct {
	h   ehandler.Order
	bus broker.Bus[entity.EventOrderCreated]
}

func NewOrderCreatedEventsCatcher(
	h ehandler.Order,
	bus broker.Bus[entity.EventOrderCreated],
) processor.Processor {
	return &orderCreatedProc{h, bus}
}

func (p *orderCreatedProc) StartAsync(ctx context.Context, wg *sync.WaitGroup) {
	const S1 = "Не удалось подписаться на топик"
	const S2 = "Подписка на топик инициализирована"

	err := p.bus.Subscribe(ctx, wg, p.h.CallbackOrderCreated)
	if err != nil {
		log.Fatal().Err(err).Str("topic_name", "order.created").Msg(S1)
	} else {
		log.Debug().Str("topic_name", "order.created").Msg(S2)
	}
}
