package persistence

import (
	"github.com/cockroachdb/pebble"
	"io"
	"log"
)

type Storage struct {
	db *pebble.DB
}

func NewStorage(storageDir string) *Storage {
	db, err := pebble.Open(storageDir, &pebble.Options{})
	if err != nil {
		log.Fatal(err)
	}

	return &Storage{db: db}
}

func (s *Storage) Get(key []byte) ([]byte, io.Closer, error) {
	value, closer, err := s.db.Get(key)
	return value, closer, err
}

func (s *Storage) Set(key []byte, value []byte) error {
	return s.db.Set(key, value, pebble.Sync)
}

type IterOptions = pebble.IterOptions

func (s *Storage) NewIter(opts *IterOptions) (*pebble.Iterator, error) {
	return s.db.NewIter(opts)
}
