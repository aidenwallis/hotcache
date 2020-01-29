# hotcache
A small hash-map cache in Golang designed for small TTL. It's a simple hashmap implementation that allows you to effectively handle TTL on temporary keys. TTL works on last set wins, meaning if you set a longer TTL on a key, it will expire when the new TTL is set.

It invalidates keys by checking 1000 random keys in store every 100ms, as well as does an expiry check per lookup/set.

Hotcache is completely thread-safe due to its use of RWMutexes, therefore you don't need to be concerned with doing that yourself. I originally wrote this package months ago and chose to make it public to just make my life easier for [Fossabot](https://fossabot.com).

## Usage

This package is incredibly easy to use, like so

```go
package main

import (
	"fmt"

	"github.com/aidenwallis/hotcache"
)

func main() {
	cache := hotcache.New()

	cache.Set("key", "value", time.Second*2)

	value, ok := cache.Get("key")
	if ok {
		// This key exists in the hashmap.
		fmt.Println("Key value: " + value.(string))
	}

    // Sleep until cache expires.
    time.Sleep(time.Second * 2)
    
    exists := cache.Has("key")
    if !exists {
        fmt.Println("Key has now expired!")
    }
}
```
