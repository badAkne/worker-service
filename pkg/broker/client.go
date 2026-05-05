package broker

import (
	"errors"

	"github.com/IBM/sarama"
	butil "github.com/badAkne/worker-service/pkg/broker/util"
)

type KafkaConfig struct {
	Addresses     []string
	ConsumerGroup string
	ClientID      string
}

type KafkaClient struct {
	addresses            []string
	defaultConsumerGroup string
	saramaCfg            *sarama.Config
	producer             sarama.SyncProducer
}

func NewKafkaClient(cfg KafkaConfig) (*KafkaClient, error) {
	clientID := butil.Coalesce(cfg.ClientID, cfg.ConsumerGroup)
	defaultGroup := butil.Coalesce(cfg.ConsumerGroup, cfg.ClientID)

	saramaCfg := sarama.NewConfig()

	saramaCfg.ClientID = clientID
	saramaCfg.Version = sarama.V2_8_0_0
	saramaCfg.Producer.Return.Successes = true
	saramaCfg.Producer.RequiredAcks = sarama.WaitForAll
	saramaCfg.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{sarama.NewBalanceStrategyRoundRobin()}
	saramaCfg.Consumer.Return.Errors = true
	saramaCfg.Consumer.Offsets.Initial = sarama.OffsetNewest

	syncProducer, err := sarama.NewSyncProducer(cfg.Addresses, saramaCfg)
	if err != nil {
		return nil, err
	}

	return &KafkaClient{
		addresses:            cfg.Addresses,
		defaultConsumerGroup: defaultGroup,
		saramaCfg:            saramaCfg,
		producer:             syncProducer,
	}, nil
}

func (c *KafkaClient) Producer() sarama.SyncProducer {
	return c.producer
}

func (c *KafkaClient) NewConsumerGroup(groupID string) (sarama.ConsumerGroup, error) {
	return sarama.NewConsumerGroup(c.addresses, groupID, c.saramaCfg)
}

func (c *KafkaClient) DefaultConsumerGroup() string {
	return c.defaultConsumerGroup
}

func (c *KafkaClient) Close() error {
	if c.producer != nil {
		return c.producer.Close()
	}

	return errors.New("producer is nil")
}
