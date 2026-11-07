package main

import (
	"fmt"
	"sync"
)

// счетчик просмотров страницы, много горутин дергают его одновременно
var (
	views   int
	mu      sync.Mutex
	wg      sync.WaitGroup
)

func viewPage(id int) {
	defer wg.Done()
	for i := 0; i < 1000; i++ {
		mu.Lock()
		views++ // без мьютекса тут была бы гонка, потерялись бы просмотры
		mu.Unlock()
	}
	fmt.Printf("горутина %d закончила\n", id)
}

func main() {
	for i := 1; i <= 10; i++ {
		wg.Add(1)
		go viewPage(i)
	}
	wg.Wait()
	fmt.Println("всего просмотров:", views) // должно быть ровно 10000
}
