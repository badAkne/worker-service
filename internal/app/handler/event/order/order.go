package eorder

import (
	"context"
	"fmt"

	"github.com/badAkne/worker-service/internal/app/entity"
	ehandler "github.com/badAkne/worker-service/internal/app/handler/event"
	"github.com/badAkne/worker-service/internal/pkg/http/binding"
	butil "github.com/badAkne/worker-service/pkg/broker/util"
	"github.com/rs/zerolog/log"
)

type handler struct{}

func NewHandler() ehandler.Order {
	return &handler{}
}

func (h *handler) CallbackOrderCreated(
	ctx context.Context, ev *entity.EventOrderCreated, headers map[string]string,
) error {
	log.Info().
		Ctx(ctx).
		Any("msg_body", ev).
		Any("msg_headers", headers).
		Msg("Получено событие ORDER_CREATED")

	if err := binding.OnlyValidate(ev); err != nil {
		return butil.NotCriticalError(fmt.Errorf("невалидные данные в EventOrderCreated: %w", err))
	}

	return nil
}
