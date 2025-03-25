package gotx

import (
	"context"
	"errors"
)

// Since its impossible to define methods on interfaces, Transactor is used as a wrapper
// around Transactor and brings in the default logic.
type Transactor[Tx TransactionImpl] struct {
	Beginner TxBeginner[Tx]
}

type AnyTransactor = Transactor[TransactionImpl]

func (t *Transactor[Tx]) Context(
	ctx context.Context,
	opts ...any,
) (*Transaction[Tx], error) {
	if v := ctx.Value(DefaultTxKey); v != nil {
		wtx, ok := v.(*Transaction[Tx])
		if ok {
			return wtx.Unowned(), nil
		}
	}

	newTx, err := t.Beginner.Begin(ctx, opts...)
	if err != nil {
		return nil, errors.Join(ErrTxBegin, err)
	}

	newTx.state.ctx = context.WithValue(ctx, DefaultTxKey, newTx)
	return newTx, nil
}

func (t *Transactor[Tx]) Any() *AnyTransactor {
	return &AnyTransactor{
		Beginner: &AnyTransactorBeginner[Tx]{
			Beginner: t.Beginner,
		},
	}
}

func WrapOwnedTransaction[Tx TransactionImpl](ctx context.Context, tx Tx) *Transaction[Tx] {
	t := &Transaction[Tx]{
		tx:      tx,
		isOwner: true,
		state: &TxState{
			err: nil,
			ctx: ctx,
		},
	}

	if v := ctx.Value(DefaultTxKey); v == nil {
		ctx = context.WithValue(ctx, DefaultTxKey, t)
		t.state.ctx = ctx
	}

	return t
}
