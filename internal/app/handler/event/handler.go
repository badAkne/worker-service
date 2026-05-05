package ehandler

import (
	"context"

	"github.com/badAkne/worker-service/internal/app/entity"
)

type Order interface {
	CallbackOrderCreated(ctx context.Context, ev *entity.EventOrderCreated, headers map[string]string) error
}
