package store_test

import (
	"testing"

	badger "github.com/dgraph-io/badger/v3"
	"github.com/plural-labs/autostaker/store"
	"github.com/plural-labs/autostaker/types"
	"github.com/stretchr/testify/require"
)

func TestRecords(t *testing.T) {
	db, err := store.New(t.TempDir())
	require.NoError(t, err)

	record := &types.Record{
		Address:   "address1",
		Frequency: types.Frequency_DAILY,
		Tolerance: 1000,
	}
	_, err = db.GetRecord(record.Address)
	require.Equal(t, badger.ErrKeyNotFound, err)

	require.NoError(t, db.SetRecord(record))

	out, err := db.GetRecord(record.Address)
	require.NoError(t, err)
	require.Equal(t, record.Address, out.Address)

	require.NoError(t, db.SetRecord(&types.Record{
		Address:   "address2",
		Frequency: types.Frequency_DAILY,
		Tolerance: 5000,
	}))

	records, err := db.GetRecordsByFrequency(int32(types.Frequency_DAILY))
	require.NoError(t, err)
	require.Len(t, records, 2)

	records, err = db.GetRecordsByFrequency(int32(types.Frequency_MONTHLY))
	require.NoError(t, err)
	require.Len(t, records, 0)

	// update the frequency to weekly
	record.Frequency = types.Frequency_WEEKLY
	require.NoError(t, db.SetRecord(record))

	// This should have removed the record from the daily frequency table
	records, err = db.GetRecordsByFrequency(int32(types.Frequency_DAILY))
	require.NoError(t, err)
	require.Len(t, records, 1)

	require.NoError(t, db.DeleteRecord("address2"))

	_, err = db.GetRecord("address2")
	require.Equal(t, badger.ErrKeyNotFound, err)
}

