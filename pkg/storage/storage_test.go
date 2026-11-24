package storage_test

import (
	"bytes"
	"testing"

	"jetpack-joyride/pkg/storage"
)

func TestStorageSaveAndLoad(t *testing.T) {
	s := storage.NewMemoryStorage()

	key := "test_achievement"
	data := []byte(`{"unlocked": true, "score": 1250}`)

	err := s.Save(key, data)
	if err != nil {
		t.Fatalf("failed to save data: %v", err)
	}

	loaded, err := s.Load(key)
	if err != nil {
		t.Fatalf("failed to load data: %v", err)
	}

	if !bytes.Equal(loaded, data) {
		t.Errorf("expected %s, got %s", string(data), string(loaded))
	}

	// Non-existent key
	_, err = s.Load("non_existent")
	if err == nil {
		t.Errorf("expected error loading non-existent key, got nil")
	}
}
