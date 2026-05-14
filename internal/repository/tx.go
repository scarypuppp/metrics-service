package repository

import (
	"context"

	"github.com/jmoiron/sqlx"
)

type txKey struct{}

func WithTx(ctx context.Context, tx *sqlx.Tx) context.Context {
	return context.WithValue(ctx, txKey{}, tx)
}
