package dict

import (
	"sync"
)

type SyncDict struct {
	m sync.Map
}

func MakeSyncDict() *SyncDict {
	return &SyncDict{}
}

// get value by key
func (d *SyncDict) Get(key string) (val interface{}, exists bool) {
	if value, ok := d.m.Load(key); ok {
		return value, true
	}
	return nil, false
}

// get dict length
func (d *SyncDict) Len() int {
	count := 0
	d.m.Range(func(key, value interface{}) bool {
		count++
		return true
	})
	return count
}

// put value by key, return 1 if success, 0 if fail
func (d *SyncDict) Put(key string, val interface{}) (result int) {
	_, ok := d.m.Load(key)
	d.m.Store(key, val)
	if ok {
		return 0
	}
	return 1
}

// put value by key if absent, return 1 if success, 0 if fail
func (d *SyncDict) PutIfAbsent(key string, val interface{}) (result int) {
	_, ok := d.m.Load(key)
	if ok {
		return 0
	}
	d.m.Store(key, val) //not ok, so store the value
	return 1
}

// put value by key if exists, return 1 if success, 0 if fail
func (d *SyncDict) PutIfExists(key string, val interface{}) (result int) {
	_, ok := d.m.Load(key)
	if !ok {
		return 0
	}
	d.m.Store(key, val)
	return 1
}

// remove value by key, return 1 if success, 0 if fail
func (d *SyncDict) Remove(key string, val interface{}) (result int) {
	_, ok := d.m.Load(key)
	if !ok {
		return 0
	}
	d.m.Delete(key) //not ok, so delete the value
	return 1
}

// foreach iterate all key-value pairs
func (d *SyncDict) ForEach(consumer Consumer) {
	d.m.Range(func(key, value interface{}) bool {
		consumer(key.(string), value)
		return true
	})
}

// get all keys
func (d *SyncDict) Keys() []string {
	keys := make([]string, d.Len())
	d.m.Range(func(key, value interface{}) bool {
		keys = append(keys, key.(string))
		return true
	})
	return keys
}

// get n random keys
func (d *SyncDict) RandomKey(n int) []string {
	keys := make([]string, d.Len())
	for i := 0; i < d.Len(); i++ {
		d.m.Range(func(key, value interface{}) bool {
			keys = append(keys, key.(string))
			return true
		})
	}
	return keys
}

// get n distinct random keys
func (d *SyncDict) RandomDistinctKeys(n int) []string {
	keys := make([]string, d.Len())
	i := 0
	d.m.Range(func(key, value interface{}) bool {
		keys[i] = key.(string)
		i++
		return i != n
	})
	return keys
}

// clear all key-value pairs
func (d *SyncDict) Clear() {
	*d = *MakeSyncDict()
}
