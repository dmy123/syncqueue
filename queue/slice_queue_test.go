package queue

import (
	"context"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"testing"
	"time"
)

//func TestSliceQueue_In(t *testing.T) {
//	s := SliceQueue[int]{}
//	s.In2(context.Background(), 1)
//	s.In2(context.Background(), 2)
//	s.In2(context.Background(), 3)
//	s.In2(context.Background(), 4)
//}

func TestSliceQueue_In(t *testing.T) {
	type testCase[T any] struct {
		name     string
		s        *SliceQueue[T]
		ctx      context.Context
		in       int
		wantData error
		wantErr  error
	}
	tests := []testCase[int]{
		{
			name: "time exceed",
			ctx: func() context.Context {
				ctx, _ := context.WithTimeout(context.Background(), time.Second)
				return ctx
			}(),
			in: 10,
			s: func() *SliceQueue[int] {
				q := NewSliceQueue[int](2)
				q.In(context.Background(), 11)
				q.In(context.Background(), 12)
				return q
			}(),
			wantErr: context.DeadlineExceeded,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.s.In(tt.ctx, tt.in)
			assert.Equal(t, err, tt.wantErr)
			if err != nil {
				return
			}
			assert.Equal(t, tt.s.data, tt.wantData)
		})
	}
}

func TestSliceQueue_Out(t *testing.T) {
	type testCase[T any] struct {
		name     string
		s        *SliceQueue[T]
		ctx      context.Context
		wantData error
		wantErr  error
	}
	tests := []testCase[int]{
		{
			name: "time exceed",
			ctx: func() context.Context {
				ctx, _ := context.WithTimeout(context.Background(), time.Second)
				return ctx
			}(),
			s: func() *SliceQueue[int] {
				q := NewSliceQueue[int](2)
				q.Out(context.Background())
				return q
			}(),
			wantErr: context.DeadlineExceeded,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val, err := tt.s.Out(tt.ctx)
			assert.Equal(t, err, tt.wantErr)
			if err != nil {
				return
			}
			assert.Equal(t, val, tt.wantData)
		})
	}
}

func TestSliceQueue_InOut(t *testing.T) {
	q := NewSliceQueue[int](10)
	closed := false
	for i := 0; i < 20; i++ {
		go func() {
			for {
				if closed {
					return
				}
				val := rand.Int()
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				//err := q.In(ctx, val)
				_ = q.In(ctx, val)

				cancel()
			}
		}()
	}
	for i := 0; i < 5; i++ {
		go func() {
			for {
				if closed {
					return
				}
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				//val, err := q.Out(ctx)
				_, _ = q.Out(ctx)
				cancel()
				// 如何校验val的值
			}
		}()
	}
	time.Sleep(time.Second * 10)
	closed = true
}
