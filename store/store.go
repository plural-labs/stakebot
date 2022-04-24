package store

import (
	"path/filepath"

	badger "github.com/dgraph-io/badger/v3"
	"github.com/google/orderedcode"
	"google.golang.org/protobuf/proto"

	"github.com/plural-labs/stakebot/types"
)

const (
	defaultStoreName = "store.db"

	addressPrefix = byte(0x00)
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

func (s Store) SetRecord(record *types.Record) error {
	// if the address exists elsewhere we remove it
	err := s.DeleteRecord(record.Address)
	if err != nil {
		return err
	}
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
			if item != nil {
				break
			}
		}
		// unable to find the address at any frequency
		if err != nil {
			return err
		}

		// unmarshal value
		return item.Value(func(val []byte) error {
			return proto.Unmarshal(val, record)
		})
	})
	if err != nil {
		return nil, err
	}
	return record, nil
}

func (s Store) DeleteRecord(address string) error {
	return s.db.Update(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = false
		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Seek([]byte{addressPrefix}); it.Valid(); it.Next() {
			err := txn.Delete(it.Item().Key())
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func (s Store) GetRecordsByFrequency(frequency int32) ([]*types.Record, error) {
	records := make([]*types.Record, 0)
	err := s.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		it := txn.NewIterator(opts)
		defer it.Close()
		prefix, err := orderedcode.Append([]byte{addressPrefix}, int64(frequency))
		if err != nil {
			panic(err)
		}
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			record := new(types.Record)
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

func (s Store) Len() (int, error) {
	recordCount := 0
	err := s.db.Update(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = false
		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Seek([]byte{addressPrefix}); it.Valid(); it.Next() {
			recordCount++
		}
		return nil
	})
	return recordCount, err
}

func key(frequency int32, address string) []byte {
	key, err := orderedcode.Append([]byte{addressPrefix}, int64(frequency), address)
	if err != nil {
		panic(err)
	}
	return key
}
