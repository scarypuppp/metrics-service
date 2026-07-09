package repository

import (
	"context"

	"github.com/jmoiron/sqlx"
)

type txKey struct{}

// WithTx returns a context carrying the given SQL transaction for use by storage methods.
func WithTx(ctx context.Context, tx *sqlx.Tx) context.Context {
	return context.WithValue(ctx, txKey{}, tx)
}
