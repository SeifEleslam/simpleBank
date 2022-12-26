package db

import (
	"context"
	"fmt"
	"math"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTransferTx(t *testing.T) {
	store := NewStore(testDB)

	account1 := CreateRandomAccount(t)
	account2 := CreateRandomAccount(t)
	fmt.Println(">> before:", account1.Balance, account2.Balance)

	errs := make(chan error)
	results := make(chan TransferTxResult)

	n := 15
	amount := float64(20)

	for i := 0; i < n; i++ {
		go func() {
			result, err := store.TransferTx(context.Background(), TranferTxParams{
				FromAccountID: account1.ID,
				ToAccountID:   account2.ID,
				Amount:        amount,
			})

			errs <- err
			results <- result

		}()
	}
	existed := make(map[int]bool)
	for i := 0; i < n; i++ {
		err := <-errs
		result := <-results

		require.NoError(t, err)

		// Test Transfer
		transfer := result.Transfer
		require.NotEmpty(t, result.Transfer)
		require.Equal(t, transfer.FromAccountID, account1.ID)
		require.Equal(t, transfer.ToAccountID, account2.ID)
		require.Equal(t, transfer.Amount, amount)
		require.NotZero(t, transfer.ID)
		require.NotZero(t, transfer.CreatedAt)

		_, err = store.GetTransfer(context.Background(), transfer.ID)
		require.NoError(t, err)

		// Test FromEntry
		fromEntry := result.FromEntry
		require.NotEmpty(t, fromEntry)
		require.Equal(t, account1.ID, fromEntry.AccountID)
		require.Equal(t, -amount, fromEntry.Amount)
		require.NotZero(t, fromEntry.ID)
		require.NotZero(t, fromEntry.CreatedAt)

		_, err = store.GetEntry(context.Background(), fromEntry.ID)
		require.NoError(t, err)

		// Test ToEntry
		toEntry := result.ToEntry
		require.NotEmpty(t, toEntry)
		require.Equal(t, account2.ID, toEntry.AccountID)
		require.Equal(t, amount, toEntry.Amount)
		require.NotZero(t, toEntry.ID)
		require.NotZero(t, toEntry.CreatedAt)

		_, err = store.GetEntry(context.Background(), toEntry.ID)
		require.NoError(t, err)

		// Test FromAccount and ToAccount
		fromAccount := result.FromAccount
		require.NotEmpty(t, fromAccount)
		require.Equal(t, fromAccount.ID, account1.ID)

		toAccount := result.ToAccount
		require.NotEmpty(t, toAccount)
		require.Equal(t, toAccount.ID, account2.ID)

		// Test the validation of the amount of transacion
		fmt.Println(">> tx:", fromAccount.Balance, toAccount.Balance)
		diff1 := account1.Balance - fromAccount.Balance
		diff2 := toAccount.Balance - account2.Balance
		require.Equal(t, diff1, diff2)
		require.True(t, diff1 > 0)
		whole, frac := math.Modf(diff1 / amount)
		require.True(t, frac == 0.0)
		require.True(t, whole >= 1 && whole <= float64(n))
		require.NotContains(t, existed, int64(whole))
		existed[int(whole)] = true
	}

	updatedAccount1, err := testQueries.GetAccount(context.Background(), account1.ID)
	require.NoError(t, err)

	updatedAccount2, err := testQueries.GetAccount(context.Background(), account2.ID)
	require.NoError(t, err)

	fmt.Println(">> After:", updatedAccount1.Balance, updatedAccount2.Balance)

	require.Equal(t, account1.Balance-float64(n)*amount, updatedAccount1.Balance)
	require.Equal(t, account2.Balance+float64(n)*amount, updatedAccount2.Balance)

}

func TestTransferTxDeadlock(t *testing.T) {
	store := NewStore(testDB)

	account1 := CreateRandomAccount(t)
	account2 := CreateRandomAccount(t)
	fmt.Println(">> before:", account1.Balance, account2.Balance)

	errs := make(chan error)

	n := 10
	amount := float64(20)

	for i := 0; i < n; i++ {
		FromAccountID := account1.ID
		ToAccountID := account2.ID

		if i%2 == 0 {
			FromAccountID = account2.ID
			ToAccountID = account1.ID
		}

		go func() {
			_, err := store.TransferTx(context.Background(), TranferTxParams{
				FromAccountID: FromAccountID,
				ToAccountID:   ToAccountID,
				Amount:        amount,
			})

			errs <- err

		}()
	}
	for i := 0; i < n; i++ {
		err := <-errs
		require.NoError(t, err)
	}

	updatedAccount1, err := testQueries.GetAccount(context.Background(), account1.ID)
	require.NoError(t, err)

	updatedAccount2, err := testQueries.GetAccount(context.Background(), account2.ID)
	require.NoError(t, err)

	fmt.Println(">> After:", updatedAccount1.Balance, updatedAccount2.Balance)
	if n%2 == 0 {
		require.Equal(t, account1.Balance, updatedAccount1.Balance)
		require.Equal(t, account2.Balance, updatedAccount2.Balance)
	} else {
		require.Equal(t, account1.Balance, updatedAccount1.Balance-amount)
		require.Equal(t, account2.Balance, updatedAccount2.Balance+amount)
	}

}
