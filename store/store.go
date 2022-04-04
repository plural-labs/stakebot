package store

import (
	"fmt"
	"path/filepath"

	sdk "github.com/cosmos/cosmos-sdk/types"
	badger "github.com/dgraph-io/badger/v3"
	"google.golang.org/protobuf/proto"

	"github.com/plural-labs/autostaker/types"
)

const defaultStoreName = "store.db"

const addressPrefix = byte(0x00)

type Store struct {
	db *badger.DB

	// immutable
	chains []types.Chain
}

func New(dir string, chains []types.Chain) (*Store, error) {
	path := filepath.Join(dir, defaultStoreName)
	db, err := badger.Open(badger.DefaultOptions(path))
	if err != nil {
		return nil, err
	}
	return &Store{
		db: db,
		chains: chains,
	}, nil
}

func (s Store) Set(record *types.Record) error {
	return s.db.Update(func(txn *badger.Txn) error {
		k, err := key(record.Address)
		if err != nil {
			return err
		}
		bz, err := proto.Marshal(record)
		if err != nil {
			return err
		}
		return txn.Set(k, bz)
	})	
}

func (s Store) Get(address string) (*types.Record, error) {
	var record *types.Record
	err := s.db.View(func(txn *badger.Txn) error {
		k, err := key(address)
		if err != nil {
			return err
		}
		item, err := txn.Get(k)
		if err != nil {
			return err
		}
		item.Value(func (val []byte) error {
			return proto.Unmarshal(val, record)
		})
		return nil
	})
	if err != nil {
		return nil, err
	}
	return record, nil
}

func (s Store) GetAll() ([]*types.Record, error) {
	records := make([]*types.Record, 0)
	err := s.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Seek([]byte{addressPrefix}); it.ValidForPrefix([]byte{addressPrefix}); it.Next() {
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

func key(address string) ([]byte, error) {
	a, err := sdk.AccAddressFromBech32(address)
	if err != nil {
		return nil, fmt.Errorf("unable to parse address: %w", err)
	}
	return append([]byte{addressPrefix}, a...), nil
}
