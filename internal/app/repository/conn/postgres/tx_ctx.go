package rcpostgres

import (
	"context"

	"github.com/uptrace/bun"
)

type _ctxKeyTx struct{}

// getTxFromContext извлекает bun.Tx из переданного контекста.
// Возвращает пустой Tx, если контекст не содержит транзакцию.
func getTxFromContext(ctx context.Context) bun.Tx {
	tx, _ := ctx.Value(_ctxKeyTx{}).(bun.Tx)
	return tx
}

// setTxToContext возвращает новый context.Context с сохранённой bun.Tx.
func setTxToContext(ctx context.Context, tx bun.Tx) context.Context {
	return context.WithValue(ctx, _ctxKeyTx{}, tx)
}
