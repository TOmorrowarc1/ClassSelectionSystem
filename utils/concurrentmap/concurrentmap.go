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
	data map[K]V
	lock sync.RWMutex
}

func NewConcurrentMap[K comparable, V any]() *ConcurrentMap[K, V] {
	return &ConcurrentMap[K, V]{
		data: make(map[K]V),
		lock: sync.RWMutex{},
	}
}

func (it *ConcurrentMap[K, V]) ReadPair(key K) (V, bool) {
	it.lock.RLock()
	defer it.lock.RUnlock()
	value, ok := it.data[key]
	return value, ok
}

func (it *ConcurrentMap[K, V]) WritePair(key K, value *V) {
	it.lock.Lock()
	defer it.lock.Unlock()
	it.data[key] = *value
}

func (it *ConcurrentMap[K, V]) DeletePair(key K) {
	it.lock.Lock()
	defer it.lock.Unlock()
	delete(it.data, key)
}

func (it *ConcurrentMap[K, V]) Load(fileName string) error {
	it.lock.Lock()
	defer it.lock.Unlock()
	file, err := os.OpenFile(fileName, os.O_RDONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&it.data)
	if err != nil {
		return err
	}
	return nil
}

func (it *ConcurrentMap[K, V]) Store(fileName string) error {
	it.lock.RLock()
	defer it.lock.RUnlock()
	file, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer file.Close()
	encoder := json.NewEncoder(file)
	err = encoder.Encode(&it.data)
	if err != nil {
		return err
	}
	return nil
}

func (it *ConcurrentMap[K, V]) Clear() {
	it.lock.Lock()
	defer it.lock.Unlock()
	it.data = make(map[K]V)
}

func (it *ConcurrentMap[K, V]) ReadAll() map[K]V {
	it.lock.RLock()
	defer it.lock.RUnlock()
	newMap := make(map[K]V)
	for k, v := range it.data {
		newMap[k] = v
	}
	return newMap
}
