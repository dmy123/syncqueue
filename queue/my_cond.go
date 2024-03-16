package queue

import (
	"context"
	"sync"
)

type Cond struct {
	cond *sync.Cond
}

func NewCond(mu *sync.RWMutex) *Cond {
	return &Cond{cond: sync.NewCond(mu)}
}

func (c Cond) Wait(ctx context.Context) {
	//ch := make(chan struct{}, 1)
	//
	//go func() {
	//	c.cond.Wait()
	//	c.cond.L.Unlock()
	//	ch <- struct{}{}
	//}()
	ch := make(chan struct{})

	go func() {
		c.cond.Wait()
		c.cond.L.Unlock()
		select {
		case ch <- struct{}{}:
		default:
			c.cond.L.Lock()
			c.cond.Signal()
			c.cond.L.Unlock()
		}
	}()
	c.cond.L.Unlock()
	select {
	case <-ctx.Done():
		return
	case <-ch:
		return
	}
}
