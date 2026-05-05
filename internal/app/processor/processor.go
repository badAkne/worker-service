package processor

import (
	"context"
	"sync"
)

// Processor — интерфейс для всех процессоров (HTTP, Kafka, Cron и т.д.).
// Каждый процессор запускается асинхронно и управляется через context.
type Processor interface {
	// StartAsync запускает процессор в отдельной горутине.
	// ctx — контекст для graceful shutdown (при отмене процессор должен завершиться).
	// wg — WaitGroup для ожидания завершения всех процессоров.
	StartAsync(ctx context.Context, wg *sync.WaitGroup)
}

// ProcessorFunc — адаптер для использования функции как Processor.
type ProcessorFunc func(ctx context.Context, wg *sync.WaitGroup)

func (p ProcessorFunc) StartAsync(ctx context.Context, wg *sync.WaitGroup) {
	p(ctx, wg)
}
