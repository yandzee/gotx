package gotx

import (
	"context"
	"log/slog"
)

type InMemoryTransactor = Transactor[*InMemoryTransaction]

type InMemTxBeginner struct {
	Log        *slog.Logger
	BeginError error
	TxError    error
}

func (imb *InMemTxBeginner) Begin(
	ctx context.Context,
	opts ...any,
) (*Transaction[*InMemoryTransaction], error) {
	attrs := []any{
		slog.Any("opts", opts),
	}

	if imb.BeginError != nil {
		attrs = append(attrs, slog.String("begin-err", imb.BeginError.Error()))
	}

	if imb.TxError != nil {
		attrs = append(attrs, slog.String("tx-err", imb.TxError.Error()))
	}

	imb.log().Debug("txClient.Begin()", attrs...)

	if imb.BeginError != nil {
		return nil, imb.BeginError
	}

	return WrapOwnedTransaction(ctx, &InMemoryTransaction{
		Log:   imb.log().With("module", "InMemoryTransaction"),
		Error: imb.TxError,
	}), nil
}

func (imb *InMemTxBeginner) log() *slog.Logger {
	if imb.Log != nil {
		return imb.Log
	}

	imb.Log = discardLogger()
	return imb.Log
}
