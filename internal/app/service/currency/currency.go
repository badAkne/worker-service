package scurrency

import (
	"context"
	"errors"

	"github.com/badAkne/worker-service/internal/app/client/fixer"
	"github.com/badAkne/worker-service/internal/app/entity"
	"github.com/badAkne/worker-service/internal/app/repository"
)

type Service struct {
	fixerClient *fixer.Client
	rateRepo    repository.CurrencyRate
}

func NewService(fixerClient *fixer.Client, rateRepo repository.CurrencyRate) *Service {
	return &Service{
		fixerClient: fixerClient,
		rateRepo:    rateRepo,
	}
}

func (s *Service) GetRate(ctx context.Context, from, to string) (float64, error) {
	if from == to {
		return 1.0, nil
	}

	rate, err := s.rateRepo.GetRate(ctx, from, to)
	if errors.Is(err, nil) {
		return rate, nil
	}

	rates, err := s.fixerClient.GetRates(ctx, from)
	if err != nil {
		return -1, err
	}

	err = s.rateRepo.SetRates(ctx, from, rates)
	if err != nil {
		return -1, err
	}

	rate, ok := rates[to]
	if !ok {
		return 0, entity.ErrFixerCurrencyNotFound
	}

	return rate, nil
}

func (s *Service) Convert(ctx context.Context, amount float64, from, to string) (float64, error) {
	rate, err := s.GetRate(ctx, from, to)
	if err != nil {
		return 0, err
	}
	return amount * rate, nil
}
