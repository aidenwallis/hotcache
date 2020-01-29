package hotcache

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGetNonexistent(t *testing.T) {
	cache := New()
	defer cache.Stop()
	val, ok := cache.Get("xd")
	assert.Equal(t, val, nil)
	assert.Equal(t, ok, false)
}

func TestGetExists(t *testing.T) {
	cache := New()
	defer cache.Stop()
	cache.Set("xd", "xd", 0)
	val, ok := cache.Get("xd")
	assert.Equal(t, val, "xd")
	assert.Equal(t, ok, true)
}

func TestExpiry(t *testing.T) {
	cache := New()
	defer cache.Stop()

	cache.Set("xd", "xd", time.Millisecond*10)

	val, ok := cache.Get("xd")
	assert.Equal(t, val, "xd")
	assert.Equal(t, ok, true)

	time.Sleep(time.Millisecond * 10)

	val, ok = cache.Get("xd")
	assert.Equal(t, val, nil)
	assert.Equal(t, ok, false)
}

func TestDelete(t *testing.T) {
	cache := New()
	defer cache.Stop()

	cache.Set("xd", "xd", 0)

	val, ok := cache.Get("xd")
	assert.Equal(t, val, "xd")
	assert.Equal(t, ok, true)

	cache.Delete("xd")

	val, ok = cache.Get("xd")
	assert.Equal(t, val, nil)
	assert.Equal(t, ok, false)
}

func TestHas(t *testing.T) {
	cache := New()
	defer cache.Stop()

	cache.Set("xd", "xd", 0)

	exists := cache.Has("xd")
	assert.Equal(t, exists, true)
}

func TestHasExpiry(t *testing.T) {
	cache := New()
	defer cache.Stop()

	cache.Set("xd", "xd", time.Millisecond*10)

	exists := cache.Has("xd")
	assert.Equal(t, exists, true)

	time.Sleep(time.Millisecond * 10)

	exists = cache.Has("xd")
	assert.Equal(t, exists, false)
}

func TestSetNX(t *testing.T) {
	cache := New()
	defer cache.Stop()

	set := cache.SetNX("xd", "xd", 0)
	assert.Equal(t, set, true)

	val, ok := cache.Get("xd")
	assert.Equal(t, val, "xd")
	assert.Equal(t, ok, true)

	set = cache.SetNX("xd", "xd", 0)
	assert.Equal(t, set, false)
}

func TestSetNXExpiry(t *testing.T) {
	cache := New()
	defer cache.Stop()

	set := cache.SetNX("xd", "xd", time.Millisecond*10)
	assert.Equal(t, set, true)

	val, ok := cache.Get("xd")
	assert.Equal(t, val, "xd")
	assert.Equal(t, ok, true)

	set = cache.SetNX("xd", "xd", 0)
	assert.Equal(t, set, false)

	time.Sleep(time.Millisecond * 10)

	set = cache.SetNX("xd", "xd2", 0)
	assert.Equal(t, set, true)

	val, ok = cache.Get("xd")
	assert.Equal(t, val, "xd2")
	assert.Equal(t, ok, true)
}
