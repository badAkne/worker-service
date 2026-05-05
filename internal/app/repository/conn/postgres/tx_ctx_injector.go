package rcpostgres

import (
	"context"
	"database/sql"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/schema"
)

// implBunIdbTxInjector реализует bun.IDB, при этом, во всех методах,
// пытается сначала найти bun.Tx в переданном контексте.
// В случае неудачи, использует fallback.
type implBunIdbTxInjector struct {
	fallback bun.IDB
}

// newBunIdbTxInjector оборачивает переданный bun.IDB таким образом,
// возвращая другую реализацию bun.IDB, которая будет уважать потенциальное
// наличие контекста транзакции, и выполнять запросы через эту транзакцию,
// в случае её наличия.
func newBunIdbTxInjector(orig bun.IDB) bun.IDB {
	return &implBunIdbTxInjector{orig}
}

func (x *implBunIdbTxInjector) QueryContext(
	ctx context.Context,
	query string, args ...any,
) (*sql.Rows, error) {
	return x.getIDB(ctx).QueryContext(ctx, query, args...)
}

func (x *implBunIdbTxInjector) ExecContext(
	ctx context.Context,
	query string, args ...any,
) (sql.Result, error) {
	return x.getIDB(ctx).ExecContext(ctx, query, args...)
}

func (x *implBunIdbTxInjector) QueryRowContext(
	ctx context.Context, query string, args ...any,
) *sql.Row {
	return x.getIDB(ctx).QueryRowContext(ctx, query, args...)
}

func (x *implBunIdbTxInjector) Dialect() schema.Dialect {
	return x.fallback.Dialect()
}

func (x *implBunIdbTxInjector) NewValues(model any) *bun.ValuesQuery {
	return x.fallback.NewValues(model).Conn(x)
}

func (x *implBunIdbTxInjector) NewSelect() *bun.SelectQuery {
	return x.fallback.NewSelect().Conn(x)
}

func (x *implBunIdbTxInjector) NewInsert() *bun.InsertQuery {
	return x.fallback.NewInsert().Conn(x)
}

func (x *implBunIdbTxInjector) NewUpdate() *bun.UpdateQuery {
	return x.fallback.NewUpdate().Conn(x)
}

func (x *implBunIdbTxInjector) NewDelete() *bun.DeleteQuery {
	return x.fallback.NewDelete().Conn(x)
}

func (x *implBunIdbTxInjector) NewMerge() *bun.MergeQuery {
	return x.fallback.NewMerge().Conn(x)
}

func (x *implBunIdbTxInjector) NewRaw(
	query string, args ...any,
) *bun.RawQuery {
	return x.fallback.NewRaw(query, args...).Conn(x)
}

func (x *implBunIdbTxInjector) NewCreateTable() *bun.CreateTableQuery {
	return x.fallback.NewCreateTable().Conn(x)
}

func (x *implBunIdbTxInjector) NewDropTable() *bun.DropTableQuery {
	return x.fallback.NewDropTable().Conn(x)
}

func (x *implBunIdbTxInjector) NewCreateIndex() *bun.CreateIndexQuery {
	return x.fallback.NewCreateIndex().Conn(x)
}

func (x *implBunIdbTxInjector) NewDropIndex() *bun.DropIndexQuery {
	return x.fallback.NewDropIndex().Conn(x)
}

func (x *implBunIdbTxInjector) NewTruncateTable() *bun.TruncateTableQuery {
	return x.fallback.NewTruncateTable().Conn(x)
}

func (x *implBunIdbTxInjector) NewAddColumn() *bun.AddColumnQuery {
	return x.fallback.NewAddColumn().Conn(x)
}

func (x *implBunIdbTxInjector) NewDropColumn() *bun.DropColumnQuery {
	return x.fallback.NewDropColumn().Conn(x)
}

func (x *implBunIdbTxInjector) BeginTx(
	ctx context.Context, opts *sql.TxOptions,
) (bun.Tx, error) {
	return x.getIDB(ctx).BeginTx(ctx, opts)
}

func (x *implBunIdbTxInjector) RunInTx(
	ctx context.Context,
	opts *sql.TxOptions, f func(ctx context.Context, tx bun.Tx) error,
) error {
	return x.getIDB(ctx).RunInTx(ctx, opts, f)
}

////////////////////////////////////////////////////////////////////////////////
///// PRIVATE METHODS //////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

// getIDB возвращает интерфейс bun.IDB, под которым ЛИБО транзакция,
// извлечённая из контекста, либо fallback.
func (x *implBunIdbTxInjector) getIDB(ctx context.Context) bun.IDB {
	tx := getTxFromContext(ctx)
	if tx.Tx != nil {
		return tx
	}
	return x.fallback
}
