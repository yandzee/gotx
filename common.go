package gotx

import (
	"context"
	"io"
	"log/slog"
)

type contextKey struct {
	name string
}

var DefaultTxKey contextKey = contextKey{"tx"}

type TxBeginner[Tx TransactionImpl] interface {
	Begin(ctx context.Context, opts ...any) (*Transaction[Tx], error)
}

func discardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{}))
}
