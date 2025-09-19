package concurrentmap

import (
	"encoding/json"
	"os"
	"sync"
)

/*
ConcurrentMap is a thread-safe map with generic support, automatically saving/loading
to/from a specific file.
*/

type ConcurrentMap[K comparable, V any] struct {
	data_ map[K]V
	lock_ sync.RWMutex
}

func NewConcurrentMap[K comparable, V any]() *ConcurrentMap[K, V] {
	return &ConcurrentMap[K, V]{
		data_: make(map[K]V),
		lock_: sync.RWMutex{},
	}
}

func (it *ConcurrentMap[K, V]) ReadPair(key K) (V, bool) {
	it.lock_.RLock()
	defer it.lock_.RUnlock()
	value, ok := it.data_[key]
	return value, ok
}

func (it *ConcurrentMap[K, V]) WritePair(key K, value *V) {
	it.lock_.Lock()
	defer it.lock_.Unlock()
	it.data_[key] = *value
}

func (it *ConcurrentMap[K, V]) DeletePair(key K) {
	it.lock_.Lock()
	defer it.lock_.Unlock()
	delete(it.data_, key)
}

func (it *ConcurrentMap[K, V]) Load(file_name string) error {
	it.lock_.Lock()
	defer it.lock_.Unlock()
	file, err := os.OpenFile(file_name, os.O_RDONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&it.data_)
	if err != nil {
		return err
	}
	return nil
}

func (it *ConcurrentMap[K, V]) Store(file_name string) error {
	it.lock_.RLock()
	defer it.lock_.RUnlock()
	file, err := os.OpenFile(file_name, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer file.Close()
	encoder := json.NewEncoder(file)
	err = encoder.Encode(&it.data_)
	if err != nil {
		return err
	}
	return nil
}

func (it *ConcurrentMap[K, V]) Clear() {
	it.lock_.Lock()
	defer it.lock_.Unlock()
	it.data_ = make(map[K]V)
}

func (it *ConcurrentMap[K, V]) ReadAll() map[K]V {
	it.lock_.RLock()
	defer it.lock_.RUnlock()
	new_map := make(map[K]V)
	for k, v := range it.data_ {
		new_map[k] = v
	}
	return new_map
}
