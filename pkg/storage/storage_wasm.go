//go:build js && wasm

package storage

import (
	"encoding/base64"
	"syscall/js"
)

type LocalStorage struct {
	storage js.Value
}

func GetStorage() Storage {
	defer func() {
		_ = recover()
	}()
	window := js.Global()
	if !window.Get("localStorage").Truthy() {
		return NewMemoryStorage()
	}
	return &LocalStorage{storage: window.Get("localStorage")}
}

func (l *LocalStorage) Save(key string, data []byte) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = ErrNotFound
		}
	}()
	b64 := base64.StdEncoding.EncodeToString(data)
	l.storage.Call("setItem", "jetpack_"+key, b64)
	return nil
}

func (l *LocalStorage) Load(key string) (result []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = ErrNotFound
		}
	}()
	item := l.storage.Call("getItem", "jetpack_"+key)
	if !item.Truthy() {
		return nil, ErrNotFound
	}
	b64 := item.String()
	return base64.StdEncoding.DecodeString(b64)
}
