package persistence

import (
	"errors"
	"github.com/cockroachdb/pebble"
	"io"
	"log"
)

//TODO The storage part can actually be designed as interface to support multiple different storages

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

func (s *Storage) Get(key []byte) ([]byte, io.Closer, bool, error) {
	value, closer, err := s.db.Get(key)
	if errors.Is(err, pebble.ErrNotFound) {
		return value, closer, false, nil
	} else if err != nil {
		return value, closer, false, err
	}
	return value, closer, true, nil
}

func (s *Storage) Set(key []byte, value []byte) error {
	return s.db.Set(key, value, pebble.Sync)
}

func (s *Storage) Delete(key []byte) error {
	return s.db.Delete(key, pebble.Sync)
}

type IterOptions = pebble.IterOptions

func (s *Storage) NewIter(opts *IterOptions) (*pebble.Iterator, error) {
	return s.db.NewIter(opts)
}
