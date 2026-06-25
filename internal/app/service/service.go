package service

import (
	"context"

	"github.com/badAkne/worker-service/internal/app/entity"
)

type (
	// Currency — интерфейс сервиса конвертации валют.
	Currency interface {
		GetRate(ctx context.Context, from, to string) (float64, error)
		Convert(ctx context.Context, amount float64, from, to string) (float64, error)
	}

	// Delivery — интерфейс сервиса расчёта стоимости доставки.
	Delivery interface {
		CalculateDeliveryPrice(ctx context.Context, order *entity.EventOrderCreated) (*entity.EventOrderDeliveryCalculated, error)
	}
)
