package queue

import (
	"context"
)

type Queue[T any] interface {
	In(ctx context.Context, val T)
	Out(ctx context.Context) (T, error)
	IsEmpty() bool
	Clear() error
	Size() int
}

//type data[T any] struct {
//	timeout time.Duration
//}
//
//func (q *data[T]) Push(msg T) error {
//
//}
//
//func (q *data[T]) Pop() (msg T, err error) {
//
//}
