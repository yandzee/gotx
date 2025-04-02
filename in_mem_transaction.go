package gotx

import (
	"context"
	"log/slog"
)

type InMemoryTransaction struct {
	Log   *slog.Logger
	Error error
}

func (imtx *InMemoryTransaction) Commit(_ context.Context) error {
	attrs := []any{}

	if imtx.Error != nil {
		attrs = append(attrs, slog.String("err", imtx.Error.Error()))
	}

	imtx.log().Debug("tx.Commit()", attrs...)
	return imtx.Error
}

func (imtx *InMemoryTransaction) Rollback(_ context.Context) error {
	attrs := []any{}

	if imtx.Error != nil {
		attrs = append(attrs, slog.String("err", imtx.Error.Error()))
	}

	imtx.log().Debug("tx.Rollback()", attrs...)
	return imtx.Error
}

func (imtx *InMemoryTransaction) log() *slog.Logger {
	if imtx.Log != nil {
		return imtx.Log
	}

	imtx.Log = discardLogger()
	return imtx.Log
}
