package main

import (
	"fmt"
	"sync"
	"time"
)

// очередь задач: добавляют одни горутины, забирают другие,
// чтение меняет очередь (задача удаляется), поэтому все под мьютексом
type TaskQueue struct {
	mu    sync.Mutex
	tasks []string
}

func (q *TaskQueue) Add(task string) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.tasks = append(q.tasks, task)
}

// забрать задачу, если пусто вернуть false
func (q *TaskQueue) Get() (string, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()
	if len(q.tasks) == 0 {
		return "", false
	}
	task := q.tasks[0]
	q.tasks = q.tasks[1:] // убираем из списка
	return task, true
}

func main() {
	q := TaskQueue{}
	var wg sync.WaitGroup

	// producer'ы кидают задачи
	for i := 1; i <= 3; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			for j := 1; j <= 5; j++ {
				task := fmt.Sprintf("задача %d-%d", n, j)
				q.Add(task)
				fmt.Println("добавили:", task)
			}
		}(i)
	}

	// consumer'ы разбирают
	for i := 1; i <= 2; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			for {
				task, ok := q.Get()
				if !ok {
					time.Sleep(100 * time.Millisecond) // вдруг еще добавят
					continue
				}
				fmt.Printf("воркер %d взял: %s\n", n, task)
				time.Sleep(50 * time.Millisecond)
			}
		}(i)
	}

	time.Sleep(3 * time.Second) // даем поработать (в реале нужны контексты и т.д.)
	fmt.Println("осталось задач в очереди:", len(q.tasks))
}
