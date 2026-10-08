package utils

import (
	"context"
	"iter"
	"sync"

	"go.uber.org/atomic"
)

type Slot struct {
	Concurrency int

	ctx    context.Context
	cancel context.CancelFunc

	queue   chan func()
	started atomic.Bool
	wg      sync.WaitGroup
}

func (s *Slot) Close(noWait ...bool) {
	wait := len(noWait) > 0 && noWait[0]
	if s.started.CompareAndSwap(true, false) {
		close(s.queue)
		if !wait {
			s.cancel()
		}
	}
	if wait {
		s.wg.Wait()
	}
}

func (s *Slot) Do(f func()) {
	if s.isConcurrent() {
		s.startWorkers()
		s.enqueue(f)
	} else {
		f()
	}
}

func (s *Slot) DoNow(f func()) {
	s.wg.Go(f)
}

func (s *Slot) Map[T any](f func(T), params []T) {
	if !s.isConcurrent() || len(params) < 1 {
		for _, x := range params {
			f(x)
		}
		return
	}

	s.startWorkers()

	var wg sync.WaitGroup
	for _, p := range params {
		wg.Add(1)
		if !s.enqueue(func() { f(p); wg.Done() }) {
			wg.Done()
			break
		}
	}
	wg.Wait()
}

func (s *Slot) MapValue[P, R any](f func(P) R, params []P, fifo bool) iter.Seq[R] {
	if !s.isConcurrent() || len(params) < 1 {
		return func(yield func(R) bool) {
			for _, x := range params {
				if !yield(f(x)) {
					return
				}
			}
		}
	}

	s.startWorkers()

	if fifo {
		return s.mapFifo(f, params)
	}
	return s.mapValue(f, params)
}

func (s *Slot) enqueue(f func()) bool {
	select {
	case s.queue <- f:
		return true
	case <-s.ctx.Done():
		return false
	}
}

func (s *Slot) isConcurrent() bool {
	return s.Concurrency > 1
}

func (s *Slot) mapFifo[P, R any](f func(P) R, params []P) iter.Seq[R] {
	return func(yield func(R) bool) {
		results := make([]chan R, len(params))
		for i := range params {
			results[i] = make(chan R, 1)
		}
		s.wg.Go(func() {
			for i, p := range params {
				r := results[i]
				w := func() {
					defer close(r)
					select {
					case r <- f(p):
					case <-s.ctx.Done():
					}
				}
				if !s.enqueue(w) {
					for _, r := range results[i:] {
						close(r)
					}
					return
				}
			}
		})

		for _, r := range results {
			select {
			case v := <-r:
				if !yield(v) {
					return
				}
			case <-s.ctx.Done():
				return
			}
		}
	}
}

func (s *Slot) mapValue[P, R any](f func(P) R, params []P) iter.Seq[R] {
	return func(yield func(R) bool) {
		result := make(chan R, s.Concurrency)
		s.wg.Go(func() {
			var wg sync.WaitGroup
			for _, p := range params {
				w := func() {
					defer wg.Done()
					select {
					case result <- f(p):
					case <-s.ctx.Done():
					}
				}
				wg.Add(1)
				if !s.enqueue(w) {
					wg.Done()
					break
				}
			}
			wg.Wait()
			close(result)
		})

		for {
			select {
			case v, ok := <-result:
				if !ok {
					return
				}
				if !yield(v) {
					return
				}
			case <-s.ctx.Done():
				return
			}
		}
	}
}

func (s *Slot) startWorkers() {
	if s.started.CompareAndSwap(false, true) {
		s.ctx, s.cancel = context.WithCancel(context.Background())
		s.queue = make(chan func(), s.Concurrency*2)

		for range s.Concurrency {
			s.wg.Go(func() {
				for {
					select {
					case f, ok := <-s.queue:
						if !ok {
							return
						}
						f()
					case <-s.ctx.Done():
						return
					}
				}
			})
		}
	}
}

func NewSlot(concurrency int) *Slot {
	s := Slot{Concurrency: concurrency}
	return &s
}
