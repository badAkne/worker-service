package rcurrency

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/badAkne/worker-service/internal/app/repository"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

const rateKeyPrefix = "rates:"

var _ repository.CurrencyRate = (*RedisRepository)(nil)

type RedisRepository struct {
	client   *redis.Client
	cacheTTL time.Duration
}

func NewRedisRepository(client *redis.Client, cacheTTL time.Duration) *RedisRepository {
	return &RedisRepository{
		client:   client,
		cacheTTL: cacheTTL,
	}
}

func (r *RedisRepository) buildKey(from, to string) string {
	return fmt.Sprintf("%s%s:%s", rateKeyPrefix, from, to)
}

func (r *RedisRepository) GetRate(ctx context.Context, from, to string) (float64, error) {
	key := r.buildKey(from, to)

	val, err := r.client.Get(ctx, key).Result()
	if err != nil {
		return -1, fmt.Errorf("unable to get currency: %w", err)
	}

	res, err := strconv.ParseFloat(val, 64)
	if err != nil {
		return -1, fmt.Errorf("unable to parse currency: %w", err)
	}

	log.Info().Float64("val:", res).Msg("successfully get currency")

	return res, nil
}

func (r *RedisRepository) SetRate(ctx context.Context, from, to string, rate float64) error {
	key := r.buildKey(from, to)

	val := strconv.FormatFloat(rate, 'f', -1, 64)

	err := r.client.Set(ctx, key, val, r.cacheTTL).Err()
	if err != nil {
		return fmt.Errorf("unable to set value: %w", err)
	}

	return nil
}

func (r *RedisRepository) SetRates(ctx context.Context, from string, rates map[string]float64) error {
	pipe := r.client.Pipeline()

	for k, v := range rates {
		key := r.buildKey(from, k)
		pipe.Set(ctx, key, v, r.cacheTTL)
	}

	cmds, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("unable to execute pipeline: %w", err)
	}

	var successfulCount int
	for _, cmd := range cmds {
		if cmd.Err() == nil {
			successfulCount++
		}
	}

	log.Info().Int("successful count", successfulCount).Msg("executed cmds in redis")

	return nil
}
