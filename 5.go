package main

import (
	"fmt"
	"sync"
	"time"
)

// Вариант 1: чтение НЕ обновляет время 
type TTLCache1 struct {
	mu      sync.RWMutex
	data    map[string]string
	created map[string]time.Time // когда записали
	ttl     time.Duration
}

func NewTTLCache1(ttl time.Duration) *TTLCache1 {
	return &TTLCache1{
		data: make(map[string]string), created: make(map[string]time.Time), ttl: ttl,
	}
}

func (c *TTLCache1) Set(k, v string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data[k] = v
	c.created[k] = time.Now()
}

func (c *TTLCache1) Get(k string) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	v, ok := c.data[k]
	return v, ok // время не трогаем
}

func (c *TTLCache1) Cleanup() {
	c.mu.Lock()
	defer c.mu.Unlock()
	now := time.Now()
	for k, t := range c.created {
		if now.Sub(t) > c.ttl {
			fmt.Println("[cache1] удаляю устаревшее:", k)
			delete(c.data, k)
			delete(c.created, k)
		}
	}
}

// Вариант 2: чтение ОБНОВЛЯЕТ время 
type TTLCache2 struct {
	mu      sync.RWMutex
	data    map[string]string
	lastUse map[string]time.Time // последний доступ (запись или чтение)
	ttl     time.Duration
}

func NewTTLCache2(ttl time.Duration) *TTLCache2 {
	return &TTLCache2{
		data: make(map[string]string), lastUse: make(map[string]time.Time), ttl: ttl,
	}
}

func (c *TTLCache2) Set(k, v string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data[k] = v
	c.lastUse[k] = time.Now()
}

func (c *TTLCache2) Get(k string) (string, bool) {
	c.mu.Lock() // тут нужен Lock а не RLock, т.к. обновляем отметку
	defer c.mu.Unlock()
	v, ok := c.data[k]
	if ok {
		c.lastUse[k] = time.Now() // продлеваем жизнь записи
	}
	return v, ok
}

func (c *TTLCache2) Cleanup() {
	c.mu.Lock()
	defer c.mu.Unlock()
	now := time.Now()
	for k, t := range c.lastUse {
		if now.Sub(t) > c.ttl {
			fmt.Println("[cache2] удаляю устаревшее:", k)
			delete(c.data, k)
			delete(c.lastUse, k)
		}
	}
}

func main() {
	c1 := NewTTLCache1(500 * time.Millisecond)
	c2 := NewTTLCache2(500 * time.Millisecond)

	c1.Set("a", "1")
	c2.Set("a", "1")

	time.Sleep(300 * time.Millisecond)
	// обращаемся к кэшу — во втором кэше запись "омолаживается"
	if v, ok := c2.Get("a"); ok {
		fmt.Println("cache2 прочитал a =", v)
	}

	time.Sleep(400 * time.Millisecond) // прошло 700мс с записи, но 400мс с чтения

	c1.Cleanup() // запись "a" протухла по времени записи -> удалится
	c2.Cleanup() // запись "a" жива, т.к. читали недавно

	_, ok1 := c1.Get("a")
	_, ok2 := c2.Get("a")
	fmt.Println("cache1: a есть?", ok1) // false
	fmt.Println("cache2: a есть?", ok2) // true

	// в реале Cleanup запускают горутиной с тикером, типа:
	// go func() { for range time.Tick(time.Second) { c.Cleanup() } }()
}
