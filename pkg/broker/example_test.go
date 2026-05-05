package broker_test

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/badAkne/worker-service/pkg/broker"
	"github.com/badAkne/worker-service/pkg/broker/codec"
)

// OrderCreatedEvent — пример события
type OrderCreatedEvent struct {
	OrderID   string  `json:"order_id"`
	Amount    float64 `json:"amount"`
	CreatedAt string  `json:"created_at"`
}

// ExampleBus демонстрирует полный цикл работы с Bus
func ExampleBus() {
	// 1. Создаём клиент Kafka
	consumerGroup := "order-service"

	client, err := broker.NewKafkaClient(broker.KafkaConfig{
		Addresses:     []string{"localhost:9092"},
		ConsumerGroup: consumerGroup,
		ClientID:      "", // опционально; для Sarama подставится ConsumerGroup
	})
	if err != nil {
		fmt.Printf("Failed to create client: %v\n", err)
		return
	}
	defer client.Close()

	// 2. Создаём Bus для топика orders.created
	// Пустая строка — NewBus подставит client.DefaultConsumerGroup()
	ordersBus := broker.MustKafkaBus[OrderCreatedEvent](
		client,
		codec.NewCodecJson[OrderCreatedEvent](),
		"orders.created",
		"",
	)
	defer ordersBus.Close()

	// 3. Отправляем сообщение
	event := &OrderCreatedEvent{
		OrderID:   "order-123",
		Amount:    99.99,
		CreatedAt: time.Now().Format(time.RFC3339),
	}

	if err := ordersBus.Send(context.Background(), event); err != nil {
		fmt.Printf("Failed to send: %v\n", err)
		return
	}

	fmt.Println("Message sent successfully")

	// 4. Подписываемся на топик (в отдельной горутине)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var wg sync.WaitGroup

	handler := func(ctx context.Context, msg *OrderCreatedEvent, headers map[string]string) error {
		fmt.Printf("Received: OrderID=%s, Amount=%.2f\n", msg.OrderID, msg.Amount)
		return nil
	}

	if err := ordersBus.Subscribe(ctx, &wg, handler); err != nil {
		fmt.Printf("Failed to subscribe: %v\n", err)
		return
	}

	wg.Wait()
	fmt.Println("Subscription completed")
}

// ExampleBusMock демонстрирует использование мока в тестах
func ExampleBusMock() {
	// Создаём мок без Kafka
	mockBus := broker.NewBusMock[OrderCreatedEvent]("orders.created")

	// Отправляем сообщение
	event := &OrderCreatedEvent{
		OrderID: "order-456",
		Amount:  150.00,
	}

	mockBus.Send(context.Background(), event)

	// Проверяем отправленные сообщения
	messages := mockBus.GetSentMessages()
	if len(messages) > 0 {
		fmt.Printf("Sent %d messages\n", len(messages))
		fmt.Printf("First message: OrderID=%s\n", messages[0].Msg.OrderID)
	}

	// Output:
	// Sent 1 messages
	// First message: OrderID=order-456
}
