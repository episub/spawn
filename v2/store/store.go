package store

import (
	"context"
	"fmt"
	"sync"
)

// DataStore Provides a threadsafe way to use a MAP across threads.  Good for use in a context as a way of passing values back up, and sharing across threads
type DataStore struct {
	data map[string]interface{}
	s    *sync.RWMutex
}

// NewDataStore Returns a new initialised data store
func NewDataStore() DataStore {
	s := sync.RWMutex{}
	return DataStore{data: make(map[string]interface{}), s: &s}
}

// AddValue Add value to store
func (d *DataStore) AddValue(name string, value interface{}) {
	d.s.Lock()
	d.data[name] = value
	d.s.Unlock()
}

// ContextAddValue Convenience function to add value to a store contained in context
func ContextAddValue(ctx context.Context, store string, name string, value interface{}) error {
	d, ok := ctx.Value(store).(DataStore)
	if !ok {
		return fmt.Errorf("Store '%s' not found in context", store)
	}

	d.AddValue(name, value)
	return nil
}

// ReadValue Reads value from the store
func (d *DataStore) ReadValue(name string) interface{} {
	var v interface{}
	d.s.RLock()
	v = d.data[name]
	d.s.RUnlock()

	return v
}

// ContextReadValue Convenience function to retrieve value stored in a store stored in context
func ContextReadValue(ctx context.Context, store string, name string) (interface{}, error) {
	d, ok := ctx.Value(store).(DataStore)
	if !ok {
		return nil, fmt.Errorf("Store '%s' not found in context", store)
	}

	return d.ReadValue(name), nil
}

// MergedMap Provide a map to this function, and it will copy all values from the datastore into that map, so it can be used in a separate thread safely
func (d *DataStore) MergedMap(newMap map[string]interface{}) {
	d.s.Lock()
	for k, v := range d.data {
		newMap[k] = v
	}
	d.s.Unlock()
}
