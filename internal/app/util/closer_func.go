package util

import (
	"context"
	"time"
)

// CloserFunc — адаптер для использования функции как io.Closer.
type CloserFunc func() error

// CloserContextFunc — функция закрытия с контекстом.
type CloserContextFunc = func(ctx context.Context) error

// CloserFuncNoErr — функция закрытия без возврата ошибки.
type CloserFuncNoErr func()

func (f CloserFunc) Close() error {
	return f()
}

func (f CloserFuncNoErr) Close() error {
	f()
	return nil
}

// CallIgnoreError вызывает функцию и игнорирует ошибку.
func CallIgnoreError(f CloserFunc) {
	_ = f()
}

// NewCloserContextFunc создаёт CloserFunc из функции с контекстом и таймаутом.
// Используется для graceful shutdown с таймаутом.
//
// Пример:
//
//	closer := util.NewCloserContextFunc(server.Shutdown, context.Background(), 5*time.Second)
//	go processor.WatchForShutdown(ctx, wg, closer)
func NewCloserContextFunc(
	f CloserContextFunc,
	ctx context.Context, timeout time.Duration,
) CloserFunc {
	return func() error {
		if timeout > 0 {
			var cancelFunc func()
			ctx, cancelFunc = context.WithTimeout(ctx, timeout)
			defer cancelFunc()
		}
		return f(ctx)
	}
}
