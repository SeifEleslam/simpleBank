package db

import (
	"context"
	"database/sql"
	"fmt"
)

type Store struct {
	*Queries
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{
		db:      db,
		Queries: New(db),
	}
}

func (store *Store) execTx(ctx context.Context, fn func(*Queries) error) error {
	tx, err := store.db.BeginTx(ctx, nil)

	if err != nil {
		return err
	}

	q := New(tx)
	err = fn(q)
	if err != nil {
		if rberr := tx.Rollback(); rberr != nil {
			return fmt.Errorf("tx err: %v, rb err: %v", err, rberr)
		}
		return err
	}
	return tx.Commit()
}

type TranferTxParams struct {
	FromAccountID int64   `json:"from_account_id"`
	ToAccountID   int64   `json:"to_account_id"`
	Amount        float64 `json:"amount"`
}

type TransferTxResult struct {
	Transfer    Transfer `json:"transfer"`
	FromAccount Account  `json:"from_account"`
	ToAccount   Account  `json:"to_account"`
	FromEntry   Entry    `json:"from_entry"`
	ToEntry     Entry    `json:"to_entry"`
}

func (store *Store) TransferTx(ctx context.Context, arg TranferTxParams) (TransferTxResult, error) {
	var result TransferTxResult

	err := store.execTx(ctx, func(q *Queries) error {
		var err error

		// Create Transfer
		result.Transfer, err = q.CreateTransfer(ctx, CreateTransferParams{
			FromAccountID: arg.FromAccountID,
			ToAccountID:   arg.ToAccountID,
			Amount:        float64(arg.Amount),
		})
		if err != nil {
			return err
		}

		// Create FromEntry
		result.FromEntry, err = q.CreateEntry(ctx, CreateEntryParams{
			AccountID: arg.FromAccountID,
			Amount:    -float64(arg.Amount),
		})
		if err != nil {
			return err
		}

		// Create ToEntry
		result.ToEntry, err = q.CreateEntry(ctx, CreateEntryParams{
			AccountID: arg.ToAccountID,
			Amount:    float64(arg.Amount),
		})
		if err != nil {
			return err
		}

		// TODO: update accounts

		// Update FromAccount
		if arg.FromAccountID < arg.ToAccountID {
			result.FromAccount, result.ToAccount, err = moneyTransaction(ctx, q, arg.FromAccountID, arg.ToAccountID, -arg.Amount, arg.Amount)
		} else {
			result.FromAccount, result.ToAccount, err = moneyTransaction(ctx, q, arg.ToAccountID, arg.FromAccountID, arg.Amount, -arg.Amount)
		}
		if err != nil {
			return err
		}

		return nil
	})
	return result, err
}

func moneyTransaction(
	ctx context.Context,
	q *Queries,
	account1ID int64,
	account2ID int64,
	amount1 float64,
	amount2 float64,
) (account1 Account, account2 Account, err error) {
	account1, err = q.AddToAccountBalance(ctx, AddToAccountBalanceParams{
		Amount: amount1,
		ID:     account1ID,
	})
	if err != nil {
		return
	}
	account2, err = q.AddToAccountBalance(ctx, AddToAccountBalanceParams{
		Amount: amount2,
		ID:     account2ID,
	})
	return

}
