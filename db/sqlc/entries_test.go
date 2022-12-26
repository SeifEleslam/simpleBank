package db

import (
	"context"
	"testing"
	"time"

	"github.com/simplebank/db/util"
	"github.com/stretchr/testify/require"
)

func CreateRandomEntry(t *testing.T) Entry {
	account := GetRandomAccount(t)
	arg := CreateEntryParams{
		AccountID: account.ID,
		Amount:    float64(util.RandomMoney()),
	}
	entry, err := testQueries.CreateEntry(context.Background(), arg)

	require.NoError(t, err)
	require.NotEmpty(t, entry)

	require.Equal(t, account.ID, entry.AccountID)
	require.Equal(t, arg.Amount, entry.Amount)

	return entry
}

func TestCreateEntry(t *testing.T) {
	CreateRandomEntry(t)
}

func TestGetEntry(t *testing.T) {
	entry1 := CreateRandomEntry(t)

	entry2, err := testQueries.GetEntry(context.Background(), entry1.ID)

	require.NoError(t, err)
	require.NotEmpty(t, entry2)

	require.Equal(t, entry1.AccountID, entry2.AccountID)
	require.Equal(t, entry1.Amount, entry2.Amount)
	require.Equal(t, entry1.ID, entry2.ID)
	require.Equal(t, entry1.CreatedAt, entry2.CreatedAt)
	require.WithinDuration(t, entry1.CreatedAt, entry2.CreatedAt, time.Second)

}

func TestListEntries(t *testing.T) {
	for i := 0; i < 10; i++ {
		CreateRandomEntry(t)
	}
	arg := ListEntriesParams{
		Limit:  5,
		Offset: 5,
	}
	entries, err := testQueries.ListEntries(context.Background(), arg)

	require.Len(t, entries, 5)
	require.NoError(t, err)
	for _, entry := range entries {
		require.NotEmpty(t, entry)
	}
}

func TestDeleteEntry(t *testing.T) {
	entry := CreateRandomEntry(t)
	err := testQueries.DeleteEntry(context.Background(), entry.ID)

	require.NoError(t, err)

	entry, err = testQueries.GetEntry(context.Background(), entry.ID)

	require.Error(t, err)
	require.Empty(t, entry)

}
