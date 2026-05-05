package pprocessor

import (
	"context"
	"sync"

	"github.com/rs/zerolog/log"

	"github.com/badAkne/worker-service/internal/app/processor"
	"github.com/badAkne/worker-service/internal/app/repository"
)

type procMigrate struct {
	migrator repository.Migrate
}

// NewMigrator создаёт процессор для выполнения миграций БД.
func NewMigrator(migrator repository.Migrate) processor.Processor {
	return &procMigrate{migrator}
}

func (p *procMigrate) StartAsync(ctx context.Context, wg *sync.WaitGroup) {
	processor.Wrap(ctx, wg, p.job)
}

////////////////////////////////////////////////////////////////////////////////
///// PRIVATE METHODS //////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

func (p *procMigrate) job(ctx context.Context) {
	oldVer, newVer, err := p.migrator.Migrate(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Ошибка при выполнении миграций")
		return
	}

	if oldVer != newVer {
		log.Info().Int64("old_ver", oldVer).Int64("new_ver", newVer).
			Msg("Схема БД обновлена")
	} else {
		log.Info().Msg("Схема БД актуальна, нечего мигрировать")
	}
}
