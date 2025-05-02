package pool

import (
	"fmt"
	"sync"
	"time"
)

var (
	SessionPool       *RoutinePool
	HandlerWorkerPool *RoutinePool
)

// Task 需要处理的任务
type Task struct {
	ID  int
	Job func() error
}

// RoutinePool 协程池
type RoutinePool struct {
	TaskQueue  chan Task
	numWorkers int
	wg         sync.WaitGroup // 等待所有任务完成
	closeOnce  sync.Once      // 保证只关闭一次
	closedChan chan struct{}  // 用于优雅退出
	timeout    time.Duration  // 设置超时时间
}

// NewPool 创建协程池
func NewPool(numWorkers, queueSize int) *RoutinePool {
	return &RoutinePool{
		TaskQueue:  make(chan Task, queueSize),
		numWorkers: numWorkers,
		closedChan: make(chan struct{}),
		timeout:    5 * time.Second,
	}
}

// Run 启动协程池
func (p *RoutinePool) Run() {
	for i := 0; i < p.numWorkers; i++ {
		go p.startWorker()
	}
}

func (p *RoutinePool) startWorker() {
	for {
		select {
		case task, ok := <-p.TaskQueue:
			if !ok {
				// 通道被关闭了，安全退出
				return
			}
			p.wg.Add(1)
			// fmt.Printf("Worker is processing task %d\n", task.ID)
			if err := task.Job(); err != nil {
				fmt.Printf("Error processing task %d: %v\n", task.ID, err)
			}
			p.wg.Done()
		case <-p.closedChan:
			// 收到关闭信号，退出
			return
		}
	}
}

// AddTask 添加任务
func (p *RoutinePool) AddTask(task Task) {
	p.TaskQueue <- task
}

// Shutdown 优雅关闭协程池
func (p *RoutinePool) Shutdown() {
	p.closeOnce.Do(func() {
		close(p.closedChan) // 通知所有worker退出
		// 等待任务完成
		done := make(chan struct{})
		go func() {
			p.wg.Wait()
			close(done)
		}()

		select {
		case <-done:
			close(p.TaskQueue) // 完成任务，关闭队列
		case <-time.After(p.timeout):
			fmt.Println("Timeout reached during shutdown, force exit.")
			close(p.TaskQueue) // 超时，强制关闭
		}
	})
}

func init() {
	HandlerWorkerPool = NewPool(5, 10)
	HandlerWorkerPool.Run()

	SessionPool = NewPool(5, 10)
	SessionPool.Run()
}
