package eorder

import (
	"context"
	"fmt"

	"github.com/badAkne/worker-service/internal/app/entity"
	ehandler "github.com/badAkne/worker-service/internal/app/handler/event"
	"github.com/badAkne/worker-service/internal/app/service"
	"github.com/badAkne/worker-service/internal/pkg/http/binding"
	"github.com/badAkne/worker-service/pkg/broker"
	butil "github.com/badAkne/worker-service/pkg/broker/util"
	"github.com/rs/zerolog/log"
)

type handler struct {
	deliveryService            service.Delivery
	busOrderDeliveryCalculated broker.Bus[entity.EventOrderDeliveryCalculated]
}

func NewHandler(
	deliveryService service.Delivery,
	busOrderDeliveryCalculated broker.Bus[entity.EventOrderDeliveryCalculated],
) ehandler.Order {
	return &handler{
		deliveryService:            deliveryService,
		busOrderDeliveryCalculated: busOrderDeliveryCalculated,
	}
}

func (h *handler) CallbackOrderCreated(
	ctx context.Context,
	ev *entity.EventOrderCreated,
	headers map[string]string,
) error {
	log.Info().
		Ctx(ctx).
		Str("order_id", ev.OrderID).
		Str("currency", ev.Currency).
		Float64("total_amount", ev.TotalAmount).
		Msg("Получено событие ORDER_CREATED")

	// Валидация
	if err := binding.OnlyValidate(ev); err != nil {
		return butil.NotCriticalError(fmt.Errorf("невалидные данные в EventOrderCreated: %w", err))
	}

	// Расчёт стоимости доставки
	deliveryEvent, err := h.deliveryService.CalculateDeliveryPrice(ctx, ev)
	if err != nil {
		return fmt.Errorf("failed to calculate delivery price: %w", err)
	}

	// Отправка события ORDER_DELIVERY_CALCULATED
	if err := h.busOrderDeliveryCalculated.Send(ctx, deliveryEvent); err != nil {
		return fmt.Errorf("failed to send ORDER_DELIVERY_CALCULATED: %w", err)
	}

	log.Info().
		Ctx(ctx).
		Str("order_id", ev.OrderID).
		Float64("delivery_price", deliveryEvent.DeliveryPrice).
		Str("currency", deliveryEvent.Currency).
		Msg("Событие ORDER_DELIVERY_CALCULATED отправлено")

	return nil
}
