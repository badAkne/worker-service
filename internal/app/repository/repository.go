package repository

import (
	"context"
)

type (
	// Migrate — интерфейс для выполнения миграций.
	Migrate interface {
		Migrate(ctx context.Context) (oldVer, newVer int64, err error)
	}

	// Transactional — интерфейс для выполнения кода внутри транзакции.
	// Репозитории, которые поддерживают транзакции, должны реализовывать этот интерфейс.
	Transactional interface {
		InsideTx(ctx context.Context, cb func(ctx context.Context) error) error
	}

	// Closer — интерфейс для закрытия подключения.
	Closer interface {
		Close() error
	}

	// TODO: Добавьте интерфейсы репозиториев для вашего проекта.
	// Пример:
	//
	// User interface {
	//     Transactional
	//
	//     GetByID(ctx context.Context, id uint32) (entity.User, error)
	//     GetByEmail(ctx context.Context, email string) (entity.User, error)
	//     Create(ctx context.Context, user entity.User) error
	//     Update(ctx context.Context, user entity.User) error
	//     Delete(ctx context.Context, id uint32) error
	// }
)
