package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/badAkne/worker-service/internal/app/client/fixer"
	"github.com/badAkne/worker-service/internal/app/config"
	rcurrency "github.com/badAkne/worker-service/internal/app/repository/currency"
	scurrency "github.com/badAkne/worker-service/internal/app/service/currency"
	"github.com/redis/go-redis/v9"
)

func main() {
	ctx := context.Background()
	args := config.LoadArgs{
		Output: os.Stdout,
	}
	// Загрузка конфигурации
	config.Load(args)
	cfg := config.Root
	// Подключение к Redis
	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6380",
	})
	defer func() {
		err := redisClient.Close()
		log.Fatal(err)
	}()

	// Инициализация компонентов
	fixerClient := fixer.NewClient(cfg.Client.Fixer)
	currencyRepo := rcurrency.NewRedisRepository(redisClient, cfg.Client.Fixer.CacheTTL)
	currencyService := scurrency.NewService(fixerClient, currencyRepo)

	// Тест 1: Запрос курса (API)
	log.Println("--- Тест 1: EUR -> USD (API) ---")
	start := time.Now()
	rate, err := currencyService.GetRate(ctx, "EUR", "USD")
	if err != nil {
		fmt.Printf("Ошибка: %v", err)
		panic("m")
	}
	log.Printf("EUR -> USD: %.6f (took %v)\n", rate, time.Since(start))

	// Тест 2: Запрос курса (Cache)
	log.Println("--- Тест 2: EUR -> USD (Cache) ---")
	start = time.Now()
	rate, err = currencyService.GetRate(ctx, "EUR", "USD")
	if err != nil {
		log.Fatalf("Ошибка: %v", err)
	}
	log.Printf("EUR -> USD: %.6f (took %v)\n", rate, time.Since(start))

	log.Println("✓ Все тесты пройдены!")
}
