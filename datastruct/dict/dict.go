package dict
// function type for iterating over key-value pairs
type Consumer func(key string, val interface{}) bool

type Dict interface {
	Get(key string) (val interface{}, ok bool)            //get value by key
	Len() int                                             //get dict length
	Put(key string, val interface{}) (result int)         //put value by key, return 1 if success, 0 if fail
	PutIfAbsent(key string, val interface{}) (result int) //put value by key if absent, return 1 if success, 0 if fail
	PutIfExists(key string, val interface{}) (result int) //put value by key if exists, return 1 if success, 0 if fail
	Remove(key string, val interface{}) (result int)      //remove value by key, return 1 if success, 0 if fail
	ForEach(consumer Consumer)                            // iterate all key-value pairs
	Keys() []string                                       // get all keys
	RandomKey(n int) []string                             // get n random keys
	RandomDistinctKeys(n int) []string                    // get n distinct random keys
	Clear()                                                // clear all key-value pairs
}
