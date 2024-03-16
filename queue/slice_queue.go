package queue

import (
	"context"
	"golang.org/x/sync/semaphore"
	"sync"
)

//// 基于切片的实现
//type SliceQueue[T any] struct {
//	data []T
//	zero T
//	mu sync.RWMutex
//}
//
//func (s SliceQueue[T]) In(ctx context.Context, val T) {
//	s.mu.Lock()
//	s.data = append(s.data, val)
//	s.mu.Lock()
//}
//
//func (s SliceQueue[T]) Out(ctx context.Context) (T, error) {
//	s.mu.Lock()
//	if s.IsEmpty(){
//		s.mu.Lock()
//		panic("queue is empty")
//	}
//	res := s.data[0]
//	// 释放元素可以垃圾回收掉。
//	s.data[0] = s.zero
//	s.data = s.data[1:]
//	s.mu.Lock()
//	return res, nil
//}
//
//func (s SliceQueue[T]) IsEmpty() bool {
//	return len(s.data) == 0
//}
//
//func (s SliceQueue[T]) Clear() error {
//	s.data = make([]T, 0)
//	return nil
//}
//
//func (s SliceQueue[T]) Size() int {
//	return len(s.data)
//}

//// 基于ring buffer的实现
//type SliceQueue[T any] struct {
//	data     []T
//	capacity int
//	front    int
//	end      int
//	size     int
//	zero     T
//	mu       *sync.RWMutex
//	//dequeue  *sync.Cond
//	//enqueue *sync.Cond
//	dequeue  *Cond
//	enqueue *Cond
//}
//
//func NewSliceQueue[T any](capacity int) *SliceQueue[T] {
//	mu := &sync.RWMutex{}
//	return &SliceQueue[T]{
//		data:     make([]T, capacity),
//		capacity: capacity,
//		mu:       mu,
//		dequeue:  NewCond(mu),
//		enqueue: NewCond(mu),
//	}
//}
//
//func (s SliceQueue[T]) In(ctx context.Context, val T) error {
//	s.dequeue.cond.L.Lock()
//	defer s.dequeue.cond.L.Unlock()
//	// 为什么用for不用if？
//	for s.isFull() {
//		// 队列full等待，有入队唤醒
//
//		select {
//		case <-ctx.Done():
//			return errors.New("")
//		default:
//			s.dequeue.cond.Wait()
//		}
//	}
//	s.data[s.end] = val
//	s.end = (s.end + 1) % s.capacity
//	s.size++
//	//s.enqueue.Signal()
//	s.enqueue.cond.Broadcast()
//	return nil
//}
//
//func (s SliceQueue[T]) Out(ctx context.Context) (T, error) {
//	s.enqueue.cond.L.Lock()
//	defer s.enqueue.cond.L.Unlock()
//	for s.isEmpty() {
//		// 队列empty等待，有入队唤醒
//		s.enqueue.cond.Wait()
//	}
//	res := s.data[s.front]
//	s.data[s.front] = s.zero
//	s.front = (s.front + 1) % s.capacity
//	//s.dequeue.Signal()
//	s.dequeue.cond.Broadcast()
//	s.size--
//	return res, nil
//}
//
//func (s SliceQueue[T]) IsEmpty() bool {
//	s.mu.RLock()
//	defer s.mu.RUnlock()
//	return s.isEmpty()
//}
//
//func (s SliceQueue[T]) isEmpty() bool {
//	return s.size == 0
//}
//
//func (s SliceQueue[T]) isFull() bool {
//	return s.size == s.capacity
//}
//

// 并发基于semeaphore的实现
type SliceQueue[T any] struct {
	data     []T
	capacity int
	front    int
	end      int
	size     int
	zero     T
	mu       *sync.RWMutex
	enqueue  *semaphore.Weighted
	dequeue  *semaphore.Weighted
}

func NewSliceQueue[T any](capacity int) *SliceQueue[T] {
	mu := &sync.RWMutex{}
	res := &SliceQueue[T]{
		data:     make([]T, capacity),
		capacity: capacity,
		mu:       mu,
		enqueue:  semaphore.NewWeighted(int64(capacity)),
		dequeue:  semaphore.NewWeighted(int64(capacity)),
		//dequeue:  semaphore.NewWeighted(0),                // 不能为0，会报错
	}
	_ = res.dequeue.Acquire(context.TODO(), int64(capacity))
	return res
}

func (s SliceQueue[T]) In(ctx context.Context, val T) error {
	err := s.enqueue.Acquire(ctx, 1)
	if err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	if ctx.Err() != nil {
		s.enqueue.Release(1)
		return ctx.Err()
	}

	s.data[s.end] = val
	s.end = (s.end + 1) % s.capacity
	s.size++
	s.dequeue.Release(1)
	return nil
}

func (s SliceQueue[T]) Out(ctx context.Context) (T, error) {
	s.dequeue.TryAcquire(1)
	s.mu.Lock()
	defer s.mu.Unlock()
	if ctx.Err() != nil {
		s.dequeue.Release(1)
		return s.zero, ctx.Err()
	}

	res := s.data[s.front]
	s.data[s.front] = s.zero
	s.front = (s.front + 1) % s.capacity
	//s.dequeue.Signal()
	s.size--
	s.enqueue.Release(1)
	return res, nil
}

func (s SliceQueue[T]) IsEmpty() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.isEmpty()
}

func (s SliceQueue[T]) isEmpty() bool {
	return s.size == 0
}

func (s SliceQueue[T]) isFull() bool {
	return s.size == s.capacity
}

//type SQ1 struct {
//	mu sync.Mutex
//	val []int
//}
//
//// 正确
//func (s *SQ1) Add1(v int)  {
//	s.mu.Lock()
//	s.val = append(s.val, v)
//}
//
//// 错误，多次调用会复制结构体，用的锁不是同一个
//func (s SQ1) Add2(v int)  {
//	s.mu.Lock()
//	s.val = append(s.val, v)
//}
//
//type SQ2 struct {
//	mu *sync.Mutex
//	val []int
//}
//
//// 正确
//func (s *SQ2) Add1(v int)  {
//	s.mu.Lock()
//	s.val = append(s.val, v)
//}
//
//// 正确
//func (s SQ2) Add2(v int)  {
//	s.mu.Lock()
//	s.val = append(s.val, v)
//}
