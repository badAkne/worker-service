package processor

import (
	"context"
	"io"
	"sync"
)

// WatchForShutdown ожидает отмены контекста и закрывает closer.
// Используется для graceful shutdown процессоров.
//
// Пример использования:
//
//	go WatchForShutdown(ctx, wg, listener)        // закроет listener при отмене ctx
//	go WatchForShutdown(ctx, wg, server.Shutdown) // вызовет Shutdown при отмене ctx
func WatchForShutdown(ctx context.Context, wg *sync.WaitGroup, closer io.Closer) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		<-ctx.Done()
		_ = closer.Close()
	}()
}

// Wrap оборачивает выполнение callback в горутину с поддержкой WaitGroup.
// Проверяет отмену контекста перед выполнением callback.
//
// Пример использования:
//
//	func (p *proc) StartAsync(ctx context.Context, wg *sync.WaitGroup) {
//	    processor.Wrap(ctx, wg, p.job)
//	}
func Wrap(ctx context.Context, wg *sync.WaitGroup, cb func(context.Context)) {
	if wg != nil {
		wg.Add(1)
	}

	go func() {
		if wg != nil {
			defer wg.Done()
		}
		select {
		case <-ctx.Done():
			return
		default:
			cb(ctx)
		}
	}()
}
