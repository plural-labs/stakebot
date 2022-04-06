package store

import (
	"path/filepath"

	badger "github.com/dgraph-io/badger/v3"
	"github.com/google/orderedcode"
	"google.golang.org/protobuf/proto"

	"github.com/plural-labs/autostaker/types"
)

const (
	defaultStoreName = "store.db"

	addressPrefix = byte(0x00)
	cronJobs      = byte(0x01)
)

type Store struct {
	db *badger.DB
}

func New(dir string) (*Store, error) {
	path := filepath.Join(dir, defaultStoreName)
	db, err := badger.Open(badger.DefaultOptions(path))
	if err != nil {
		return nil, err
	}
	return &Store{
		db: db,
	}, nil
}

func (s Store) SetJob(job *types.Job) error {
	return s.db.Update(func(txn *badger.Txn) error {
		bz, err := proto.Marshal(job)
		if err != nil {
			return err
		}
		return txn.Set(jobKey(int32(job.Frequency)), bz)
	})
}

func (s Store) GetJob(frequency int32) (*types.Job, error) {
	job := new(types.Job)
	err := s.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(jobKey(frequency))
		if err != nil {
			return err
		}

		// unmarshal value
		return item.Value(func(val []byte) error {
			return proto.Unmarshal(val, job)
		})
	})
	if err != nil {
		return nil, err
	}
	return job, nil
}

func (s Store) DeleteAllJobs() error {
	return s.db.Update(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		it := txn.NewIterator(opts)
		defer it.Close()
		prefix := []byte{cronJobs}
		for it.Seek(prefix); it.Valid(); it.Next() {
			item := it.Item()
			err := txn.Delete(item.Key())
			if err != nil {
				return err
			}
			return nil
		}
		return nil
	})
}

func (s Store) SetRecord(record *types.Record) error {
	return s.db.Update(func(txn *badger.Txn) error {
		bz, err := proto.Marshal(record)
		if err != nil {
			return err
		}
		return txn.Set(key(int32(record.Frequency), record.Address), bz)
	})
}

func (s Store) GetRecord(address string) (*types.Record, error) {
	record := new(types.Record)
	err := s.db.View(func(txn *badger.Txn) error {
		var (
			item *badger.Item
			err  error
		)
		// iterate over all the possible frequencies
		for frequency := int32(1); frequency <= 4; frequency++ {
			item, err = txn.Get(key(frequency, address))
			if err != nil && err != badger.ErrKeyNotFound {
				return err
			}
		}
		// unable to find the address at any frequency
		if err != nil {
			return err
		}

		// unmarshal value
		item.Value(func(val []byte) error {
			return proto.Unmarshal(val, record)
		})
		return nil
	})
	if err != nil {
		return nil, err
	}
	return record, nil
}

func (s Store) GetRecordsByFrequency(frequency int32) ([]*types.Record, error) {
	records := make([]*types.Record, 0)
	err := s.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		it := txn.NewIterator(opts)
		defer it.Close()
		prefix, err := orderedcode.Append(nil, frequency)
		if err != nil {
			panic(err)
		}
		for it.Seek(prefix); it.Valid(); it.Next() {
			var record *types.Record
			item := it.Item()
			err := item.Value(func(v []byte) error {
				if err := proto.Unmarshal(v, record); err != nil {
					return err
				}
				records = append(records, record)
				return nil
			})
			if err != nil {
				return err
			}
		}
		return nil
	})
	return records, err
}

func (s Store) Close() error {
	return s.db.Close()
}

func key(frequency int32, address string) []byte {
	key, err := orderedcode.Append([]byte{addressPrefix}, int64(frequency), address)
	if err != nil {
		panic(err)
	}
	return key
}

func jobKey(frequency int32) []byte {
	key, err := orderedcode.Append([]byte{cronJobs}, int64(frequency))
	if err != nil {
		panic(err)
	}
	return key
}
