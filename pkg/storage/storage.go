package storage

import (
	"errors"
	"sync"
)

var (
	ErrNotFound = errors.New("key not found")
)

// Storage defines the cross-platform persistence contract.
type Storage interface {
	Save(key string, data []byte) error
	Load(key string) ([]byte, error)
}

// MemoryStorage is an in-memory implementation useful for tests and fallback.
type MemoryStorage struct {
	mu   sync.RWMutex
	data map[string][]byte
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		data: make(map[string][]byte),
	}
}

func (m *MemoryStorage) Save(key string, data []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	copyData := make([]byte, len(data))
	copy(copyData, data)
	m.data[key] = copyData
	return nil
}

func (m *MemoryStorage) Load(key string) ([]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	data, ok := m.data[key]
	if !ok {
		return nil, ErrNotFound
	}
	copyData := make([]byte, len(data))
	copy(copyData, data)
	return copyData, nil
}
