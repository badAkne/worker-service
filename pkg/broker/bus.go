package broker

import (
	"context"
	"errors"
	"sync"

	"github.com/IBM/sarama"
	"github.com/badAkne/worker-service/pkg/broker/codec"
	butil "github.com/badAkne/worker-service/pkg/broker/util"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

type MessageHandler[T any] func(ctx context.Context, msg *T, headers map[string]string) error

type Bus[T any] interface {
	Send(ctx context.Context, msg *T) error
	SendWithHeaders(ctx context.Context, msg *T, headers map[string]string) error
	Subscribe(ctx context.Context, wg *sync.WaitGroup, handler MessageHandler[T]) error
	QueueName() string
	Close() error
}

type kafkaBus[T any] struct {
	client        *KafkaClient
	codec         codec.Codec[T]
	topic         string
	consumerGroup string
	consumer      sarama.ConsumerGroup
}

func NewBus[T any](client *KafkaClient, codec codec.Codec[T], topic, consumerGroup string) (Bus[T], error) {
	if client == nil {
		return nil, errors.New("client is nil")
	}

	if topic == "" {
		return nil, errors.New("topic is nil")
	}

	group := butil.Coalesce(consumerGroup, client.DefaultConsumerGroup())

	if group == "" {
		return nil, errors.New("group is nil")
	}

	return &kafkaBus[T]{
		client:        client,
		codec:         codec,
		topic:         topic,
		consumerGroup: group,
	}, nil
}

func MustKafkaBus[T any](client *KafkaClient, codec codec.Codec[T], topic, consumerGroup string) Bus[T] {
	bus, err := NewBus(client, codec, topic, consumerGroup)
	if err != nil {
		log.Fatal().Err(err).Msg("unable to start up kafka bus")
	}

	return bus
}

func (b *kafkaBus[T]) Send(ctx context.Context, msg *T) error {
	return b.SendWithHeaders(ctx, msg, nil)
}

func (b *kafkaBus[T]) SendWithHeaders(ctx context.Context, msg *T, headers map[string]string) error {
	data, err := b.codec.Encode(msg)
	if err != nil {
		return err
	}

	messageKey := uuid.New().String()

	saramaMsg := &sarama.ProducerMessage{
		Topic: b.topic,
		Key:   sarama.StringEncoder(messageKey),
		Value: sarama.ByteEncoder(data),
	}

	if headers != nil {
		saramaMsg.Headers = make([]sarama.RecordHeader, 0, len(headers))
		for k, v := range headers {
			header := sarama.RecordHeader{
				Key:   []byte(k),
				Value: []byte(v),
			}

			saramaMsg.Headers = append(saramaMsg.Headers, header)
		}
	}

	_, _, err = b.client.Producer().SendMessage(saramaMsg)
	if err != nil {
		return err
	}

	return nil
}

func (b *kafkaBus[T]) Subscribe(ctx context.Context, wg *sync.WaitGroup, handler MessageHandler[T]) error {
	consumer, err := b.client.NewConsumerGroup(b.consumerGroup)
	if err != nil {
		return err
	}

	b.consumer = consumer

	consumerHandler := consumerGroupHandler[T]{
		codec:   codec.NewCodecJson[T](),
		handler: handler,
	}
	if wg != nil {
		wg.Add(1)
	}

	go func() {
		if wg != nil {
			defer wg.Done()
		}
		defer func() {
			if err := consumer.Close(); err != nil {
				log.Error().Err(err).Msg("unable to close consumer")
			}
		}()

		for {
			if err := consumer.Consume(ctx, []string{b.topic}, &consumerHandler); err != nil {
				if ctx.Err() != nil {
					return
				}
				log.Printf("Consumer error: %v", err)
			}

			if ctx.Err() != nil {
				return
			}
		}
	}()

	return nil
}

type consumerGroupHandler[T any] struct {
	codec   codec.Codec[T]
	handler MessageHandler[T]
}

func (h *consumerGroupHandler[T]) Setup(_ sarama.ConsumerGroupSession) error {
	return nil
}

func (h *consumerGroupHandler[T]) Cleanup(_ sarama.ConsumerGroupSession) error {
	return nil
}

func (h *consumerGroupHandler[T]) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		headers := make(map[string]string, len(msg.Headers))
		for _, v := range msg.Headers {
			headers[string(v.Key)] = string(v.Value)
		}

		decoded, err := h.codec.Decode(msg.Value)
		if err != nil {
			log.Error().Err(err).Msg("unable to decode message")
			session.MarkMessage(msg, "")
			continue
		}

		ctx := session.Context()
		if err := h.handler(ctx, decoded, headers); err != nil {
			if butil.IsNotCriticalError(err) {
				session.MarkMessage(msg, "")
				continue
			}

			continue
		}

		session.MarkMessage(msg, "")
	}

	return nil
}

func (b *kafkaBus[T]) QueueName() string {
	return b.topic
}

func (b *kafkaBus[T]) Close() error {
	if b.consumer != nil {
		return b.consumer.Close()
	}

	return errors.New("consumer is nil")
}
