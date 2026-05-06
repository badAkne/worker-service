package section

type (
	// Broker - конфиг брокеров сообщений.
	Broker struct {
		Kafka BrokerKafka
	}

	// BrokerKafka - конфиг брокера сообщений Kafka.
	BrokerKafka struct {
		Addresses     []string `required:"true"`
		ConsumerGroup string   `split_words:"true"`
		ClientID      string   `split_words:"true" default:"worker-service"`

		ModelOrder BrokerKafkaModelOrder `split_words:"true"`
	}

	// BrokerKafkaModelOrder - конфиг для событий заказа.
	BrokerKafkaModelOrder struct {
		Created BrokerKafkaModelOrderCreated `split_words:"true"`
	}

	// BrokerKafkaModelOrderCreated - конфиг для топика order.created.
	// Используется: Для подписки: да, для публикации: нет.
	BrokerKafkaModelOrderCreated struct {
		Topic         string `required:"true" default:"order.created"`
		ConsumerGroup string `split_words:"true"`
	}
)
