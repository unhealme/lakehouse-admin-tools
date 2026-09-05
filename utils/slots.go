package utils

import (
	"context"
	"iter"
	"sync"
)

type Slot struct {
	Concurrency int

	sem chan EmptyType
	wg  sync.WaitGroup
}

func (s *Slot) Close() {
	s.Wait()
	if s.sem != nil {
		close(s.sem)
	}
}

func (s *Slot) Block() {
	if s.sem != nil {
		s.sem <- Empty
	}
}

func (s *Slot) Do(f func()) {
	if s.Concurrency > 0 {
		s.sem <- Empty
		s.wg.Go(func() { f(); <-s.sem })
	} else {
		f()
	}
}

func (s *Slot) DoNow(f func()) {
	if s.Concurrency > 0 {
		s.wg.Go(f)
	} else {
		f()
	}
}

func (s *Slot) Map[T any](f func(T), params []T) {
	if s.Concurrency < 1 {
		for _, x := range params {
			f(x)
		}
	}

	var wg sync.WaitGroup
	for _, x := range params {
		s.sem <- Empty
		wg.Go(func() { f(x); <-s.sem })
	}
	wg.Wait()
}

func (s *Slot) MapValue[PT, RT any](f func(PT) RT, params []PT, fifo bool) iter.Seq[RT] {
	if s.Concurrency < 1 {
		return func(yield func(RT) bool) {
			for _, x := range params {
				if !yield(f(x)) {
					return
				}
			}
		}
	}

	var (
		result  chan RT
		results []chan RT
	)
	if fifo {
		results = make([]chan RT, len(params))
		for i := range results {
			results[i] = make(chan RT, 1)
		}
	} else {
		result = make(chan RT, s.Concurrency)
	}

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		for i, x := range params {
			select {
			case <-ctx.Done():
				return
			case s.sem <- Empty:
				r := result
				if fifo {
					r = results[i]
				}
				s.wg.Go(func() { r <- f(x); <-s.sem })
			}
		}
	}()

	if fifo {
		return func(yield func(RT) bool) {
			for _, res := range results {
				if !yield(<-res) {
					cancel()
					return
				}
			}
		}
	}

	return func(yield func(RT) bool) {
		for range params {
			if !yield(<-result) {
				cancel()
				return
			}
		}
	}
}

func (s *Slot) Unblock() {
	if s.sem != nil {
		<-s.sem
	}
}

func (s *Slot) Wait() {
	if s.Concurrency > 0 {
		s.wg.Wait()
	}
}

func NewSlot(concurrency int) *Slot {
	s := Slot{Concurrency: concurrency}
	if concurrency > 0 {
		s.sem = make(chan EmptyType, concurrency)
	}
	return &s
}
