package supadb

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"sync"
)

var (
	ErrKeyNotFound    = errors.New("supadb: key not found")
	ErrClosed         = errors.New("supadb: database is closed")
	ErrFilePermission = errors.New("supadb: file permission denied")
)

type DB[K comparable, V any] struct {
	mu     sync.RWMutex
	path   string
	file   *os.File
	data   map[K]V
	closed bool
}

type Options struct {
	AutoCreate bool
	FileMode   os.FileMode
}

func Open[K comparable, V any](path string, opts Options) (*DB[K, V], error) {
	if !opts.AutoCreate {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			return nil, fmt.Errorf("file does not exist and autoCreate is false: %w", err)
		}
	}

	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, opts.FileMode)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}

	fileContent, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var data map[K]V
	if len(fileContent) == 0 {
		data = make(map[K]V)
	} else {
		err = json.Unmarshal(fileContent, &data)
		if err != nil {
			return nil, fmt.Errorf("failed to deserialize JSON: %w", err)
		}
	}

	db := DB[K, V]{
		mu:     sync.RWMutex{},
		closed: false,
		data:   data,
		file:   file,
		path:   path,
	}

	return &db, nil
}

func (db *DB[K, V]) Close() error {
	db.mu.Lock()
	defer db.mu.Unlock()

	if db.closed {
		return ErrClosed
	}

	db.closed = true

	err := db.file.Close()
	if err != nil {
		return fmt.Errorf("failed to close file: %w", err)
	}

	db.file = nil
	db.data = nil
	db.path = ""
	return nil
}

func (db *DB[K, V]) Put(key K, value V) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	if db.closed {
		return ErrClosed
	}

	db.data[key] = value
	db.save()
	return nil
}

func (db *DB[K, V]) Get(key K) (V, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	var zero V

	if db.closed {
		return zero, ErrClosed
	}

	if _, ok := db.data[key]; !ok {
		return zero, ErrKeyNotFound
	}

	value := db.data[key]
	return value, nil
}

func (db *DB[K, V]) Delete(key K) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	if db.closed {
		return ErrClosed
	}

	if _, ok := db.data[key]; !ok {
		return ErrKeyNotFound
	}

	delete(db.data, key)
	db.save()
	return nil
}

func (db *DB[K, V]) Has(key K) bool {
	db.mu.RLock()
	defer db.mu.RUnlock()

	_, ok := db.data[key]
	return ok
}

func (db *DB[K, V]) save() error {
	dataBytes, err := json.Marshal(db.data)
	if err != nil {
		return fmt.Errorf("failed to serialize data to JSON: %w", err)
	}

	err = db.file.Truncate(0)
	if err != nil {
		return fmt.Errorf("failed to truncate file: %w", err)
	}

	_, err = db.file.Seek(0, 0)
	if err != nil {
		return fmt.Errorf("failed to seek to beginning of file: %w", err)
	}

	db.file.WriteString(string(dataBytes))

	return nil
}
