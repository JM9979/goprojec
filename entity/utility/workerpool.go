package utility

import (
	"context"
	"sync"
)

// WorkerPool 表示一个通用的工作池，用于并发执行任务
type WorkerPool struct {
	maxWorkers  int                // 最大工作协程数
	taskChan    chan Task          // 任务通道
	resultsChan chan any           // 结果通道
	errorChan   chan error         // 错误通道
	wg          sync.WaitGroup     // 等待组，用于等待所有任务完成
	ctx         context.Context    // 上下文，用于取消操作
	cancel      context.CancelFunc // 取消函数
}

// Task 表示要执行的任务
type Task func() (any, error)

// NewWorkerPool 创建一个新的工作池
func NewWorkerPool(ctx context.Context, maxWorkers int, bufferSize int) *WorkerPool {
	if maxWorkers <= 0 {
		maxWorkers = 10 // 默认10个工作协程
	}

	if bufferSize <= 0 {
		bufferSize = maxWorkers * 2 // 默认缓冲区大小为工作协程数的两倍
	}

	// 创建上下文，以便可以取消操作
	ctxWithCancel, cancel := context.WithCancel(ctx)

	pool := &WorkerPool{
		maxWorkers:  maxWorkers,
		taskChan:    make(chan Task, bufferSize),
		resultsChan: make(chan any, bufferSize),
		errorChan:   make(chan error, bufferSize),
		ctx:         ctxWithCancel,
		cancel:      cancel,
	}

	// 启动工作协程
	pool.startWorkers()

	return pool
}

// startWorkers 启动工作协程池
func (p *WorkerPool) startWorkers() {
	for i := 0; i < p.maxWorkers; i++ {
		p.wg.Add(1)
		go p.worker()
	}
}

// worker 是工作协程的主函数
func (p *WorkerPool) worker() {
	defer p.wg.Done()
	for {
		select {
		case <-p.ctx.Done():
			// 上下文已取消，退出工作协程
			return
		case task, ok := <-p.taskChan:
			if !ok {
				// 任务通道已关闭，退出工作协程
				return
			}

			// 执行任务
			result, err := task()
			if err != nil {
				p.errorChan <- err
			} else if result != nil {
				p.resultsChan <- result
			}
		}
	}
}

// Submit 提交任务到工作池
func (p *WorkerPool) Submit(task Task) {
	select {
	case <-p.ctx.Done():
		// 上下文已取消，不再接受新任务
		return
	case p.taskChan <- task:
		// 任务已提交
	}
}

// GetResults 获取结果通道
func (p *WorkerPool) GetResults() <-chan any {
	return p.resultsChan
}

// GetErrors 获取错误通道
func (p *WorkerPool) GetErrors() <-chan error {
	return p.errorChan
}

// Wait 等待所有任务完成并关闭结果和错误通道
func (p *WorkerPool) Wait() {
	// 关闭任务通道，不再接受新任务
	close(p.taskChan)

	// 等待所有工作协程完成
	p.wg.Wait()

	// 关闭结果和错误通道
	close(p.resultsChan)
	close(p.errorChan)
}

// Cancel 取消工作池操作
func (p *WorkerPool) Cancel() {
	p.cancel()
}

// CollectResults 等待并收集所有结果
func (p *WorkerPool) CollectResults() ([]any, []error) {
	var results []any
	var errors []error

	// 创建一个WaitGroup来等待收集完成
	var wg sync.WaitGroup
	wg.Add(2)

	// 收集结果
	go func() {
		defer wg.Done()
		for result := range p.resultsChan {
			results = append(results, result)
		}
	}()

	// 收集错误
	go func() {
		defer wg.Done()
		for err := range p.errorChan {
			errors = append(errors, err)
		}
	}()

	// 等待任务完成并关闭通道
	p.Wait()

	// 等待收集完成
	wg.Wait()

	return results, errors
}

// WorkerPoolWithContext 使用工作池处理切片，支持类型安全的任务和结果
func WorkerPoolWithContext[T any, R any](
	ctx context.Context,
	items []T,
	maxWorkers int,
	processor func(context.Context, T) (R, error),
) ([]R, []error) {
	// 创建工作池
	pool := NewWorkerPool(ctx, maxWorkers, len(items))

	// 提交任务
	for _, item := range items {
		currentItem := item
		pool.Submit(func() (any, error) {
			result, err := processor(ctx, currentItem)
			if err != nil {
				return nil, err
			}
			return result, nil
		})
	}

	// 收集结果
	rawResults, errors := pool.CollectResults()

	// 转换结果类型
	results := make([]R, 0, len(rawResults))
	for _, raw := range rawResults {
		results = append(results, raw.(R))
	}

	return results, errors
}
