package mycache

import (
	"github.com/dgraph-io/ristretto/v2"
	"sync"
	"time"
)

var LockCache *ristretto.Cache[string, *MyLock]

func init() {
	cache, err := ristretto.NewCache[string, *MyLock](&ristretto.Config[string, *MyLock]{
		NumCounters: 300000,    // number of keys to track frequency of (10M).
		MaxCost:     200000000, // maximum cost of cache (1GB).
		BufferItems: 64,        // number of keys per Get buffer.
	})
	if err != nil {
		panic(err)
	}
	LockCache = cache
}

func GetLock(key string) *MyLock {
	LockCache.Wait()
	lock, b := LockCache.Get(key)
	if b {
		LockCache.SetWithTTL(key, lock, 0, 10*time.Minute)
		return lock
	}
	mutex := &MyLock{Lock: sync.Mutex{}}
	LockCache.SetWithTTL(key, mutex, 0, 10*time.Minute)
	LockCache.Wait()
	lock, _ = LockCache.Get(key)
	return lock
}

type MyLock struct {
	Lock sync.Mutex
}
