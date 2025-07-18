package t_pool

import (
	"fmt"
	"sync"
)

// WorkerPool 协程池结构
type WorkerPool struct {
	workerCount int
	tasks       chan func() // 任务队列
	wg          sync.WaitGroup
}

// NewWorkerPool 创建一个新的协程池
func NewWorkerPool(workerCount int) *WorkerPool {
	return &WorkerPool{
		workerCount: workerCount,
		tasks:       make(chan func()),
	}
}

// Run 启动协程池
func (wp *WorkerPool) Run() {
	for i := 0; i < wp.workerCount; i++ {
		go func(workerID int) {
			for task := range wp.tasks {
				fmt.Printf("Worker %d: executing task\n", workerID)
				task() // 执行任务
			}
		}(i + 1)
	}
}

// AddTask 添加任务到任务队列
func (wp *WorkerPool) AddTask(task func()) {
	wp.wg.Add(1)
	wp.tasks <- func() {
		defer wp.wg.Done()
		task()
	}
}

// Wait 等待所有任务完成
func (wp *WorkerPool) Wait() {
	wp.wg.Wait()
}

// Close 关闭任务队列
func (wp *WorkerPool) Close() {
	close(wp.tasks)
}
