package rcredis

import (
	"context"
	"fmt"
	"time"

	"github.com/badAkne/worker-service/internal/app/config/section"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

type Client struct {
	Cl  *redis.Client
	Cfg section.RepositoryRedis
}

func NewConn(ctx context.Context, cfg section.RepositoryRedis) (*Client, error) {
	log.Info().Str("addr", cfg.Address).Int("db", cfg.DB).Msg("Connecting to redis")

	opts := &redis.Options{
		Addr:     cfg.Address,
		Password: cfg.Password,
		DB:       cfg.DB,
	}

	cl := redis.NewClient(opts)

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	status := cl.Ping(ctx)

	if status.Err() != nil {
		return nil, fmt.Errorf("unable to ping redis: %w", status.Err())
	}

	log.Info().Msg("Connected to Redis")
	return &Client{
		Cl:  cl,
		Cfg: cfg,
	}, nil
}
