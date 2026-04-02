// go run main.go
//
// sync.RWMutex — reader-writer lock.
// Много читателей одновременно ИЛИ один писатель.
// Полезен когда читают часто, пишут редко (кэш, конфиг).

package main

import (
	"fmt"
	"sync"
	"time"
)

type Cache struct {
	mu   sync.RWMutex
	data map[string]string
}

func (c *Cache) Get(key string) (string, bool) {
	c.mu.RLock() // много читателей одновременно
	defer c.mu.RUnlock()
	v, ok := c.data[key]
	return v, ok
}

func (c *Cache) Set(key, value string) {
	c.mu.Lock() // эксклюзивный доступ
	defer c.mu.Unlock()
	c.data[key] = value
}

func main() {
	cache := &Cache{data: make(map[string]string)}
	cache.Set("hello", "world")

	var wg sync.WaitGroup

	// 10 читателей — работают параллельно.
	for i := range 10 {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			v, _ := cache.Get("hello")
			fmt.Printf("Reader %d: %s\n", id, v)
		}(i)
	}

	// 1 писатель — блокирует всех читателей.
	wg.Add(1)
	go func() {
		defer wg.Done()
		time.Sleep(time.Millisecond)
		cache.Set("hello", "updated")
		fmt.Println("Writer: updated")
	}()

	wg.Wait()
	v, _ := cache.Get("hello")
	fmt.Printf("Final: %s\n", v)
}
