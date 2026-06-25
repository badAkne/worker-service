package sdelivery

import (
	"context"
	"fmt"
	"time"

	"github.com/badAkne/worker-service/internal/app/entity"
	"github.com/badAkne/worker-service/internal/app/service"
	"github.com/rs/zerolog/log"
)

const (
	// BaseDeliveryPrice — базовая стоимость доставки в EUR.
	BaseDeliveryPrice = 10.0
	// BaseCurrency — валюта базовой стоимости.
	BaseCurrency = "EUR"
)

type deliveryService struct {
	currencyService service.Currency
}

func NewService(currencyService service.Currency) service.Delivery {
	log.Info().
		Float64("base_price", BaseDeliveryPrice).
		Str("base_currency", BaseCurrency).
		Msg("Delivery service created")

	return &deliveryService{
		currencyService: currencyService,
	}
}

func (s *deliveryService) CalculateDeliveryPrice(
	ctx context.Context,
	order *entity.EventOrderCreated,
) (*entity.EventOrderDeliveryCalculated, error) {
	// Конвертируем базовую стоимость в валюту заказа
	deliveryPrice, err := s.currencyService.Convert(ctx, BaseDeliveryPrice, BaseCurrency, order.Currency)
	if err != nil {
		return nil, fmt.Errorf("failed to convert delivery price: %w", err)
	}

	result := &entity.EventOrderDeliveryCalculated{
		OrderID:       order.OrderID,
		DeliveryPrice: deliveryPrice,
		Currency:      order.Currency,
		CalculatedAt:  time.Now().UTC().Format(time.RFC3339),
	}

	log.Info().
		Str("order_id", order.OrderID).
		Float64("base_price", BaseDeliveryPrice).
		Str("base_currency", BaseCurrency).
		Float64("delivery_price", deliveryPrice).
		Str("target_currency", order.Currency).
		Msg("Delivery price calculated")

	return result, nil
}
