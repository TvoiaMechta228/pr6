package main

import (
	"fmt"
	"sync"
	"time"
)

// кэш: читать можно много потоков сразу (RLock не блокирует),
// а писать может только один, остальные ждут
type Cache struct {
	mu sync.RWMutex
	data map[string]string
}

func NewCache() *Cache {
	return &Cache{data: make(map[string]string)}
}

func (c *Cache) Get(key string) (string, bool) {
	c.mu.RLock() // блокирующая запись, неблокирующее чтение
	defer c.mu.RUnlock()
	v, ok := c.data[key]
	return v, ok
}

func (c *Cache) Set(key, value string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	// тут типа пересобираем кэш, читатели ждут пока мы закончим
	time.Sleep(50 * time.Millisecond)
	c.data[key] = value
}

func main() {
	c := NewCache()
	var wg sync.WaitGroup

	wg.Add(1)
	go func() { // писатель
		defer wg.Done()
		for i := 0; i < 3; i++ {
			c.Set(fmt.Sprintf("key%d", i), fmt.Sprintf("value%d", i))
		}
	}()

	for i := 0; i < 5; i++ { // читатели, их несколько
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			for j := 0; j < 5; j++ {
				if v, ok := c.Get("key0"); ok {
					fmt.Printf("читатель %d: %s\n", n, v)
				} else {
					fmt.Printf("читатель %d: пока пусто\n", n)
				}
				time.Sleep(20 * time.Millisecond)
			}
		}(i)
	}
	wg.Wait()
}
