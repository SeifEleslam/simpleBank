package db

import (
	"context"
	"testing"
	"time"

	"github.com/simplebank/db/util"
	"github.com/stretchr/testify/require"
)

func CreateRandomTransfer(t *testing.T) Transfer {
	from_account := CreateRandomAccount(t)
	to_account := CreateRandomAccount(t)

	arg := CreateTransferParams{
		FromAccountID: from_account.ID,
		ToAccountID:   to_account.ID,
		Amount:        float64(util.RandomMoney()),
	}
	transfer, err := testQueries.CreateTransfer(context.Background(), arg)

	require.NoError(t, err)
	require.NotEmpty(t, transfer)

	require.Equal(t, transfer.Amount, arg.Amount)
	require.Equal(t, transfer.FromAccountID, from_account.ID)
	require.Equal(t, transfer.ToAccountID, to_account.ID)

	return transfer

}

func TestCreateTransfer(t *testing.T) {
	CreateRandomTransfer(t)
}

func TestGetTransfer(t *testing.T) {
	transfer := CreateRandomTransfer(t)
	readTransfer, err := testQueries.GetTransfer(context.Background(), transfer.ID)

	require.NoError(t, err)
	require.NotEmpty(t, readTransfer)

	require.Equal(t, transfer.Amount, readTransfer.Amount)
	require.Equal(t, transfer.ID, readTransfer.ID)
	require.Equal(t, transfer.ToAccountID, readTransfer.ToAccountID)
	require.Equal(t, transfer.FromAccountID, readTransfer.FromAccountID)
	require.Equal(t, transfer.CreatedAt, readTransfer.CreatedAt)
	require.WithinDuration(t, readTransfer.CreatedAt, transfer.CreatedAt, time.Second)

}

func TestListTransfers(t *testing.T) {
	for i := 0; i < 10; i++ {
		CreateRandomTransfer(t)
	}

	arg := ListTransfersParams{
		Limit:  5,
		Offset: 5,
	}

	transfers, err := testQueries.ListTransfers(context.Background(), arg)

	require.NoError(t, err)
	require.Len(t, transfers, 5)

	for _, transfer := range transfers {
		require.NotEmpty(t, transfer)
	}
}

func TestDeleteTransfer(t *testing.T) {
	transfer := CreateRandomTransfer(t)

	err := testQueries.DeleteTransfer(context.Background(), transfer.ID)

	require.NoError(t, err)

	transfer, err = testQueries.GetTransfer(context.Background(), transfer.ID)

	require.Error(t, err)
	require.Empty(t, transfer)
}
