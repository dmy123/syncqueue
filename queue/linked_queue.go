package queue

import (
	"errors"
	"sync/atomic"
	"unsafe"
)

// 2、基于链表的，非阻塞的，不使用锁的并发队列，基于CAS操作
type Node[T any] struct {
	tail unsafe.Pointer
	val  T
}

type LinkedQueue[T any] struct {
	head unsafe.Pointer
	tail unsafe.Pointer
	zero T
}

func NewLinkedQueue[T any]() *LinkedQueue[T] {
	return &LinkedQueue[T]{}
}

func (l *LinkedQueue[T]) In(val T) error {
	newNode := &Node[T]{val: val}
	newPtr := unsafe.Pointer(newNode)
	for {
		tailPtr := atomic.LoadPointer(&l.tail)
		tail := (*Node[T])(tailPtr)
		tailNext := atomic.LoadPointer(&tail.tail)
		if tailNext != nil {
			continue
		}
		if atomic.CompareAndSwapPointer(&tail.tail, tailNext, newPtr) {
			atomic.CompareAndSwapPointer(&l.tail, tailPtr, newPtr)
		}

	}
}

func (l *LinkedQueue[T]) Out() (T, error) {
	for {
		headPtr := atomic.LoadPointer(&l.head)
		head := (*Node[T])(headPtr)
		tailPtr := atomic.LoadPointer(&l.tail)
		tail := (*Node[T])(tailPtr)

		if head == tail {
			return l.zero, errors.New("队列为空")
		}
		headNextPtr := atomic.LoadPointer(&head.tail)
		if atomic.CompareAndSwapPointer(&l.head, headPtr, headNextPtr) {
			headNext := (*Node[T])(headNextPtr)
			return headNext.val, nil
		}
	}
}

//// 1、基于链表的，非阻塞的，使用锁的并发队列
//type Node[T any] struct {
//	front *Node[T]
//	tail  *Node[T]
//	val   T
//}
//
//type LinkedQueue[T any] struct {
//	head *Node[T]
//	end  *Node[T]
//	mu   sync.Mutex
//	zero T
//}
//
//func NewLinkedQueue[T any]() *LinkedQueue[T] {
//	return &LinkedQueue[T]{}
//}
//
//func (l *LinkedQueue[T]) In(val T) error {
//	l.mu.Lock()
//	defer l.mu.Unlock()
//	temp := &Node[T]{val: val}
//	if l.end != nil {
//		l.end.tail = temp
//	}
//	//temp.front = l.end
//	l.end = temp
//	if l.head == nil {
//		l.head = l.end
//	}
//	return nil
//}
//
//func (l *LinkedQueue[T]) Out() (T, error) {
//	if l.head == nil {
//		return l.zero, errors.New("")
//	}
//	l.mu.Lock()
//	defer l.mu.Unlock()
//	res := l.head
//	newHead := l.head.tail
//	res.tail = nil
//	//if newHead != nil {
//	//	newHead.front = nil
//	//}
//	l.head = newHead
//	if newHead == nil {
//		l.end = newHead
//	}
//
//	return res.val, nil
//}
