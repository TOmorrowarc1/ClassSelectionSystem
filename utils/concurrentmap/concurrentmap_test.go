package concurrentmap

import (
	"fmt"
	"os"
	"sync"
	"testing"
)

// Test for the concurrency safety for map
const THREAD_NUM = 100

func TestConcurrentMap_Concurrency(t *testing.T) {
	test_instance := NewConcurrentMap[string, int]()
	wg := sync.WaitGroup{}
	// First we test the reading and writing concurrently.
	for i := 0; i < THREAD_NUM; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			key := fmt.Sprintf("%s-%d", "key", index)
			value := index
			test_instance.WritePair(&key, &value)
			if read_value, ok := test_instance.ReadPair(&key); !ok || read_value != value {
				t.Errorf("Concurrency test failed at writing/reading1-key: %s, value: %d", key, value)
			}
		}(i)
	}
	wg.Wait()
	// View the result after all writing and reading are done.
	for i := 0; i < THREAD_NUM; i++ {
		key := fmt.Sprintf("%s-%d", "key", i)
		if read_value, ok := test_instance.ReadPair(&key); !ok || read_value != i {
			t.Errorf("Concurrency test failed at reading1-key: %s, value: %d", key, i)
		}
	}
	// What about conflict keys?
	for i := 0; i < THREAD_NUM; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			key := fmt.Sprintf("%s-%d", "key", index)
			value := index + 100
			test_instance.WritePair(&key, &value)
			if read_value, ok := test_instance.ReadPair(&key); !ok || read_value != value {
				t.Errorf("Concurrency test failed at writing/reading2-key: %s, value: %d", key, value)
			}
		}(i)
	}
	wg.Wait()
	// Try to read, write and erase at the same time.
	for i := 0; i < (THREAD_NUM-1)*2+1; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			switch index % 2 {
			case 0:
				key := fmt.Sprintf("%s-%d", "key", index/2)
				test_instance.DeletePair(&key)
				if _, ok := test_instance.ReadPair(&key); ok {
					t.Errorf("Concurrency test failed at deleting1-key: %s", key)
				}
			case 1:
				key := fmt.Sprintf("%s-%d", "key", index+200)
				value := index + 200
				test_instance.WritePair(&key, &value)
				if read_value, ok := test_instance.ReadPair(&key); !ok || read_value != value {
					t.Errorf("Concurrency test failed at writing/reading3-key: %s, value: %d", key, value)
				}
			}
		}(i)
	}
	wg.Wait()
}

func TestConcurrentMap_File(t *testing.T) {
	test_instance := NewConcurrentMap[string, int]()
	// First we write some data into the map.
	for i := 0; i < THREAD_NUM; i++ {
		key := fmt.Sprintf("%s-%d", "key", i)
		value := i
		test_instance.WritePair(&key, &value)
	}
	// Store it into a file.
	file_name := "test_concurrent_map.json"
	err := test_instance.Store(&file_name)
	if err != nil {
		t.Errorf("File test failed at storing: %s", err.Error())
	}
	// Clear the map.
	test_instance.Clear()
	// Load from the file.
	err = test_instance.Load(&file_name)
	if err != nil {
		t.Errorf("File test failed at loading: %s", err.Error())
	}
	// Check the result.
	for i := 0; i < THREAD_NUM; i++ {
		key := fmt.Sprintf("%s-%d", "key", i)
		if read_value, ok := test_instance.ReadPair(&key); !ok || read_value != i {
			t.Errorf("File test failed at reading-key: %s, value: %d", key, i)
		}
	}
	os.Remove(file_name)
}
