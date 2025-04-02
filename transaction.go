package gotx

import (
	"context"
	"errors"
	"sync"
)

var (
	ErrTxExhausted = errors.New("Transaction has been exhausted")
	ErrTxBegin     = errors.New("Transaction begin has failed")
	ErrTxCommit    = errors.New("Transaction commit has failed")
	ErrTxRollback  = errors.New("Transaction rollback has failed")
)

type AnyTransaction = Transaction[TransactionImpl]

type Transaction[Tx TransactionImpl] struct {
	isOwner bool
	tx      Tx
	state   *TxState
}

type TxState struct {
	sync.RWMutex

	err error
	ctx context.Context
}

type TransactionImpl interface {
	Commit(context.Context) error
	Rollback(context.Context) error
}

func (wt *Transaction[Tx]) Commit(ctx context.Context) error {
	return wt.do(func(tx Tx) error {
		err := tx.Commit(ctx)

		switch {
		case err == nil:
			return nil
		case errors.Is(err, ErrTxCommit):
			return err
		default:
			return errors.Join(err, ErrTxCommit)
		}
	})
}

func (wt *Transaction[Tx]) Rollback(ctx context.Context) error {
	return wt.do(func(tx Tx) error {
		err := tx.Rollback(ctx)

		switch {
		case err == nil:
			return nil
		case errors.Is(err, ErrTxRollback):
			return err
		default:
			return errors.Join(err, ErrTxRollback)
		}
	})
}

func (wt *Transaction[Tx]) Err() error {
	return wt.state.err
}

func (wt *Transaction[Tx]) Underlying() Tx {
	return wt.tx
}

func (wt *Transaction[Tx]) AsAnyTransaction() *AnyTransaction {
	return &Transaction[TransactionImpl]{
		isOwner: wt.isOwner,
		tx:      wt.tx,
		state:   wt.state,
	}
}

func (wt *Transaction[Tx]) Context() context.Context {
	wt.state.RLock()
	defer wt.state.RUnlock()

	return wt.state.ctx
}

func (wt *Transaction[Tx]) Unowned() *Transaction[Tx] {
	return &Transaction[Tx]{
		tx:      wt.tx,
		isOwner: false,
		state:   wt.state,
	}
}

func (wt *Transaction[Tx]) IsOwned() bool {
	return wt.isOwner
}

func (wt *Transaction[Tx]) do(fn func(Tx) error) error {
	wt.state.Lock()
	defer wt.state.Unlock()

	if wt.state.err != nil {
		return wt.state.err
	}

	if !wt.isOwner {
		return nil
	}

	err := fn(wt.tx)

	switch {
	case err == nil:
		wt.state.err = ErrTxExhausted
	case errors.Is(err, ErrTxExhausted):
		wt.state.err = err
	default:
		wt.state.err = errors.Join(err, ErrTxExhausted)
	}

	return err
}
