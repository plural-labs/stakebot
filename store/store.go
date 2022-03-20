package store

import (
	badger "github.com/dgraph-io/badger/v3"
)

type Store struct {
	db *badger.DB
}

func NewStore(dir string) (*Store, error) {
	db, err := badger.Open(badger.DefaultOptions(dir))
	if err != nil {
		return nil, err
	}
	return &Store{db: db}, nil
}

