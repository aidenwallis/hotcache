package hotcache

import (
	"math/rand"
	"sync"
	"time"
)

// cacheValue is what we nest the stored values in Hotcache with, essentially to hold metadata.
type cacheValue struct {
	expiry time.Time
	value  interface{}
}

type Hotcache struct {
	// Adds thread-safety
	expiryMutex sync.RWMutex
	storeMutex  sync.RWMutex

	// Small list of all keys that have an expiry on them, it doesn't have to be perfectly in sync as the expiry ticker will remove any redundant ones.
	expiringKeys []string

	// The actual cache store
	store map[string]*cacheValue

	// Ticker is what runs the garbage collection on a set interval.
	ticker *time.Ticker
}

func New() *Hotcache {
	h := &Hotcache{
		expiringKeys: make([]string, 0),
		store:        make(map[string]*cacheValue),
		ticker:       time.NewTicker(time.Millisecond * 100),
	}

	go h.startTicker()

	return h
}

// Stop must be called when you are done with the tempcache, as it will stop the garbage collecting ticker.
func (h *Hotcache) Stop() {
	h.ticker.Stop()

	// Clear expiry list
	h.expiryMutex.Lock()
	h.expiringKeys = make([]string, 0)
	h.expiryMutex.Unlock()

	// Clear hashmap
	h.storeMutex.Lock()
	h.store = make(map[string]*cacheValue)
	h.storeMutex.Unlock()
}

// Get retrieves a key that isn't expired from cache
func (h *Hotcache) Get(key string) (interface{}, bool) {
	h.storeMutex.RLock()
	val, ok, expired := h.get(key)
	h.storeMutex.RUnlock()

	if expired {
		h.storeMutex.Lock()
		h.evict(key)
		h.storeMutex.Unlock()
	}

	return val, ok
}

// Set adds a key to store. Use expiration of 0 for no expiry. Note this will override the key if it's existing.
func (h *Hotcache) Set(key string, value interface{}, expiration time.Duration) {
	h.storeMutex.Lock()
	if expiration != 0 {
		h.expiryMutex.Lock()
	}

	h.set(key, value, expiration)

	if expiration != 0 {
		h.expiryMutex.Unlock()
	}
	h.storeMutex.Unlock()
}

// Has checks if a key is in cache and not expired
func (h *Hotcache) Has(key string) bool {
	h.storeMutex.RLock()
	_, ok, expired := h.get(key)
	h.storeMutex.RUnlock()

	if expired {
		h.storeMutex.Lock()
		h.evict(key)
		h.storeMutex.Unlock()
	}

	return ok
}

func (h *Hotcache) Delete(key string) {
	h.storeMutex.Lock()
	delete(h.store, key)
	h.storeMutex.Unlock()
}

// get assumes that the mutex lock has already been obtained.
func (h *Hotcache) get(key string) (interface{}, bool, bool) {
	val, ok := h.store[key]

	if !ok {
		return nil, ok, false
	}

	if !val.expiry.IsZero() && val.expiry.Before(time.Now()) {
		return nil, false, true
	}

	return val.value, ok, false
}

// set assumes that the mutex lock has already been obtained.
func (h *Hotcache) set(key string, value interface{}, expiration time.Duration) {
	var expireAt time.Time
	if expiration != 0 {
		expireAt = time.Now().Add(expiration)
	}

	h.store[key] = &cacheValue{
		expiry: expireAt,
		value:  value,
	}

	if expiration != 0 {
		h.expiringKeys = append(h.expiringKeys, key)
	}
}

func (h *Hotcache) SetNX(key string, value interface{}, expiration time.Duration) bool {
	h.storeMutex.Lock()
	defer h.storeMutex.Unlock()

	_, exists, _ := h.get(key)
	if exists {
		return false
	}

	h.set(key, value, expiration)
	return true
}

// evict removes a key from cache that has expired, assumes a mutex is held
func (h *Hotcache) evict(key string) {
	// Note that we don't remove the key from h.expiringKeys, the slice is eventually consistent,
	// meaning that it's fine that the key exists in there, as randomness should eventually check the
	// key and remove it, it may not be as efficient on memory, but is far more performant than
	// performing a linear search per eviction.
	delete(h.store, key)
}

// startTicker starts the ticking process for garbage collection on it's own goroutine
func (h *Hotcache) startTicker() {
	for range h.ticker.C {
		h.tick()
	}
}

// tick is the actual tick action from the ticker that's called per interval
func (h *Hotcache) tick() {
	keylength := len(h.expiringKeys)
	if keylength == 0 {
		return
	}

	toCheck := 1000
	if keylength < toCheck {
		toCheck = keylength
	}

	// Check random keys on the expiring keys lish.
	for i := 0; i < toCheck; i++ {
		rand.Seed(time.Now().UnixNano())
		index := rand.Intn(len(h.expiringKeys))

		// Race conditions
		h.expiryMutex.RLock()
		if len(h.expiringKeys) <= index {
			// Check if key still in slice
			h.expiryMutex.RUnlock()
			continue
		}

		key := h.expiringKeys[index]
		h.expiryMutex.RUnlock()

		evicted := h.attemptEviction(key)
		if evicted {
			// Remove the key as an expiring key
			h.expiryMutex.Lock()
			h.expiringKeys[index] = h.expiringKeys[len(h.expiringKeys)-1]
			h.expiringKeys = h.expiringKeys[:len(h.expiringKeys)-1]
			h.expiryMutex.Unlock()
		}
	}
}

// attemptEviction will attempt to evict the key if it has already expired.
func (h *Hotcache) attemptEviction(key string) bool {
	h.storeMutex.RLock()
	value, ok := h.store[key]
	h.storeMutex.RUnlock()

	if !ok || value.expiry.IsZero() {
		return true // We can say it's evicted as this will never expiry anyway
	}

	if value.expiry.After(time.Now()) {
		return false
	}

	h.storeMutex.Lock()
	delete(h.store, key)
	h.storeMutex.Unlock()

	return true
}
