//go:build !js

package storage

import (
	"os"
	"path/filepath"
)

type FileStorage struct {
	dir string
}

func GetStorage() Storage {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return NewMemoryStorage()
	}
	appDir := filepath.Join(configDir, "jetpack-joyride")
	_ = os.MkdirAll(appDir, 0755)
	return &FileStorage{dir: appDir}
}

func (f *FileStorage) Save(key string, data []byte) error {
	filePath := filepath.Join(f.dir, key+".json")
	return os.WriteFile(filePath, data, 0644)
}

func (f *FileStorage) Load(key string) ([]byte, error) {
	filePath := filepath.Join(f.dir, key+".json")
	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return data, nil
}
