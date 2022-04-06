package store_test

import (
	"testing"

	badger "github.com/dgraph-io/badger/v3"
	"github.com/plural-labs/autostaker/store"
	"github.com/plural-labs/autostaker/types"
	"github.com/stretchr/testify/require"
)

func TestRecords(t *testing.T) {

}

func TestJobs(t *testing.T) {
	db, err := store.New(t.TempDir())
	require.NoError(t, err)
	job := &types.Job{
		Id: 1,
		Frequency: types.Frequency_DAILY,
	}
	_, err = db.GetJob(int32(types.Frequency_DAILY))
	require.Equal(t, badger.ErrKeyNotFound, err)

	require.NoError(t, db.SetJob(job))

	out, err := db.GetJob(int32(types.Frequency_DAILY))
	require.NoError(t, err)
	require.Equal(t, job.Id, out.Id)
	require.Equal(t, job.Frequency, out.Frequency)

	require.NoError(t, db.DeleteAllJobs())
	_, err = db.GetJob(int32(types.Frequency_DAILY))
	require.Equal(t, badger.ErrKeyNotFound, err)
}
