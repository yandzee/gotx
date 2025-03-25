package gotx

import "context"

type AnyTransactorBeginner[Tx TransactionImpl] struct {
	Beginner TxBeginner[Tx]
}

func (at *AnyTransactorBeginner[Tx]) Begin(
	ctx context.Context,
	opts ...any,
) (*AnyTransaction, error) {
	tx, err := at.Beginner.Begin(ctx, opts...)
	if err != nil {
		return nil, err
	}

	return tx.AsAnyTransaction(), nil
}
