package tests

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"testing"

	"github.com/yandzee/gotx"
)

type TestDescriptor struct {
	Name     string
	BeginErr error
	TxErr    error

	// Do Commit and then Rollback if true, vice versa if false
	CommitRollback bool
}

func TestEverything(t *testing.T) {
	runTests(t, []TestDescriptor{
		{
			Name:     "BeginFailure",
			BeginErr: errors.New("Begin test error"),
		},
		{
			Name:           "Success CommitRollback",
			BeginErr:       nil,
			TxErr:          nil,
			CommitRollback: true,
		},
		{
			Name:           "Success RollbackCommit",
			BeginErr:       nil,
			TxErr:          nil,
			CommitRollback: false,
		},
		{
			Name:           "Failed CommitRollback",
			BeginErr:       nil,
			TxErr:          errors.New("Transaction test error"),
			CommitRollback: true,
		},
		{
			Name:           "Failed RollbackCommit",
			BeginErr:       nil,
			TxErr:          errors.New("Transaction test error"),
			CommitRollback: false,
		},
	})
}

func runTests(t *testing.T, td []TestDescriptor) {
	ctx := context.Background()

	for _, d := range td {
		t.Run(d.Name, func(t *testing.T) {
			txer := newTxer(d.BeginErr, d.TxErr)

			tx, err := txer.Context(context.Background())

			switch {
			case d.BeginErr != nil && !errors.Is(err, d.BeginErr):
				t.Fatalf("txer.Context gives wrong error: %s", err.Error())
			case d.BeginErr == nil && err != nil:
				t.Fatalf("txer.Context gives unexpected error: %s", err.Error())
			case d.BeginErr != nil:
				return
			}

			if tx == nil {
				t.Fatalf("txer.Context gives nil transaction")
			}

			if !tx.IsOwned() {
				t.Fatalf("txer.Context gives unowned transaction")
			}

			// Ownership checks
			innerTx, err := txer.Context(tx.Context())
			if err != nil {
				t.Fatalf("Failed to get tx from wrapped context: %s", err.Error())
			}

			if innerTx.IsOwned() {
				t.Fatalf("Wrapped context gives owned transaction")
			}

			if err := tx.Unowned().Commit(ctx); err != nil {
				t.Fatalf("Fresh Unowned commit failed: %s\n", err.Error())
			}

			// Use checks
			if d.CommitRollback {
				err = tx.Commit(ctx)

				switch {
				case d.TxErr != nil && !errors.Is(err, d.TxErr):
					t.Fatalf("Commit gives wrong error: %s", err.Error())
				case d.TxErr == nil && err != nil:
					t.Fatalf("Commit gives unexpected error: %s", err.Error())
				case d.TxErr != nil:
					break
				}

				// Commit after exhaust check
				err = tx.Commit(ctx)
				switch {
				case err == nil:
					t.Fatalf("Repeated Commit gives no error")
				case !errors.Is(err, gotx.ErrTxExhausted):
					t.Fatalf("Repeated Commit gives error that is not ErrTxExhausted")
				case d.TxErr != nil && !errors.Is(err, gotx.ErrTxCommit):
					t.Fatalf("Repeated Commit gives error that is not ErrTxCommit: %s", err.Error())
				case d.TxErr != nil && !errors.Is(err, d.TxErr):
					t.Fatalf("Repeated Commit has lost previous error")
				}

				// Rollback on exhausted tx
				err = tx.Rollback(ctx)
				switch {
				case err == nil:
					t.Fatalf("Rollback after Commit gives no error")
				case !errors.Is(err, gotx.ErrTxExhausted):
					t.Fatalf("Rollback afer Commit gives error that is not ErrTxExhausted")
				case d.TxErr != nil && !errors.Is(err, gotx.ErrTxCommit):
					t.Fatalf("Rollback after Commit gives error that is not ErrTxCommit: %s", err.Error())
				case d.TxErr != nil && !errors.Is(err, d.TxErr):
					t.Fatalf("Rollback after Commit has lost previous error")
				}
			} else {
				err = tx.Rollback(ctx)

				switch {
				case d.TxErr != nil && !errors.Is(err, d.TxErr):
					t.Fatalf("Rollback gives wrong error: %s", err.Error())
				case d.TxErr == nil && err != nil:
					t.Fatalf("Rollback gives unexpected error: %s", err.Error())
				case d.TxErr != nil:
					break
				}

				// Rollback after exhaust check
				err = tx.Rollback(ctx)
				switch {
				case err == nil:
					t.Fatalf("Repeated Rollback gives no error")
				case !errors.Is(err, gotx.ErrTxExhausted):
					t.Fatalf("Repeated Rollback gives error that is not ErrTxExhausted")
				case d.TxErr != nil && !errors.Is(err, gotx.ErrTxRollback):
					t.Fatalf("Repeated Rollback gives error that is not ErrTxCommit: %s", err.Error())
				case d.TxErr != nil && !errors.Is(err, d.TxErr):
					t.Fatalf("Repeated Rollback has lost previous error")
				}

				// Rollback on exhausted tx
				err = tx.Commit(ctx)
				switch {
				case err == nil:
					t.Fatalf("Commit after Rollback gives no error")
				case !errors.Is(err, gotx.ErrTxExhausted):
					t.Fatalf("Commit afer Rollback gives error that is not ErrTxExhausted")
				case d.TxErr != nil && !errors.Is(err, gotx.ErrTxRollback):
					t.Fatalf("Commit after Rollback gives error that is not ErrTxCommit: %s", err.Error())
				case d.TxErr != nil && !errors.Is(err, d.TxErr):
					t.Fatalf("Commit after Rollback has lost previous error")
				}
			}
		})
	}
}

func newTxer(beginErr, txErr error) *gotx.InMemoryTransactor {
	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}))

	return &gotx.InMemoryTransactor{
		Beginner: &gotx.InMemTxBeginner{
			Log:        log,
			BeginError: beginErr,
			TxError:    txErr,
		},
	}
}
