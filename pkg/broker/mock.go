package broker

import (
	"context"
	"sync"
)

type SentMessage[T any] struct {
	Msg     *T
	Headers map[string]string
}

type BusMock[T any] struct {
	SendFunc            func(ctx context.Context, msg *T) error
	SendWithHeadersFunc func(ctx context.Context, msg *T, headers map[string]string) error
	SubscribeFunc       func(ctx context.Context, wg *sync.WaitGroup, handler MessageHandler[T]) error

	TopicValue   string
	SentMessages []SentMessage[T]
	mu           sync.Mutex
}

func NewBusMock[T any](topic string) *BusMock[T] {
	busMock := BusMock[T]{
		TopicValue:   topic,
		SentMessages: make([]SentMessage[T], 0),
	}

	return &busMock
}

func (m *BusMock[T]) Send(ctx context.Context, msg *T) error {
	return m.SendWithHeaders(ctx, msg, nil)
}

func (m *BusMock[T]) SendWithHeaders(ctx context.Context, msg *T, headers map[string]string) error {
	m.mu.Lock()

	m.SentMessages = append(m.SentMessages, SentMessage[T]{
		Msg:     msg,
		Headers: headers,
	})
	m.mu.Unlock()

	if m.SendWithHeadersFunc != nil {
		return m.SendWithHeadersFunc(ctx, msg, headers)
	}

	if m.SendFunc != nil {
		return m.SendFunc(ctx, msg)
	}

	return nil
}

func (m *BusMock[T]) Subscribe(ctx context.Context, wg *sync.WaitGroup, handler MessageHandler[T]) error {
	if m.SubscribeFunc != nil {
		return m.SubscribeFunc(ctx, wg, handler)
	}

	return nil
}

func (m *BusMock[T]) QueueName() string {
	return m.TopicValue
}

func (m *BusMock[T]) Close() error {
	return nil
}

func (m *BusMock[T]) GetSentMessages() []SentMessage[T] {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.SentMessages
}

func (m *BusMock[T]) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.SentMessages = make([]SentMessage[T], 0)
}
