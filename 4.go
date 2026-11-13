package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

// голосование в реальном времени: голоса приходят постоянно,
// запись блокирующая (мьютекс), чтобы не потерять голоса
var (
	votes = map[string]int{}
	mu    sync.Mutex
)

func voter(name string, wg *sync.WaitGroup) {
	defer wg.Done()
	candidates := []string{"Иванов", "Петров", "Сидоров"}
	for i := 0; i < 5; i++ {
		c := candidates[rand.Intn(len(candidates))]
		mu.Lock()
		votes[c]++ // подсчет голоса
		mu.Unlock()
		time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)
	}
}

func reporter(wg *sync.WaitGroup, stop chan struct{}) {
	defer wg.Done()
	ticker := time.NewTicker(300 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-stop:
			return
		case <-ticker.C:
			mu.Lock()
			fmt.Println("--- промежуточный подсчет ---")
			for c, n := range votes {
				fmt.Printf("%s: %d голосов\n", c, n)
			}
			mu.Unlock()
		}
	}
}

func main() {
	rand.Seed(time.Now().UnixNano())
	var wg sync.WaitGroup
	stop := make(chan struct{})

	wg.Add(1)
	go reporter(&wg, stop)

	for i := 1; i <= 8; i++ { // 8 избирателей голосуют параллельно
		wg.Add(1)
		go voter(fmt.Sprintf("избиратель%d", i), &wg)
	}

	// ждем пока все проголосуют
	go func() {
		wg.Add(1)
		defer wg.Done()
		time.Sleep(2 * time.Second)
	}()

	time.Sleep(2 * time.Second)
	close(stop)
	time.Sleep(100 * time.Millisecond)

	mu.Lock()
	fmt.Println("=== ИТОГИ ===")
	total := 0
	for c, n := range votes {
		fmt.Printf("%s: %d\n", c, n)
		total += n
	}
	fmt.Println("всего голосов:", total) // должно быть 8*5 = 40
	mu.Unlock()
}
