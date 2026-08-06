package utils

import (
	"context"
	"iter"
	"sync"
)

func ParallelMap[T any](task func(T), params iter.Seq[T], concurrency int) {
	sem := make(chan EmptyType, max(concurrency, 1))
	var wg sync.WaitGroup
	for i := range params {
		sem <- Empty
		wg.Go(func() { task(i); <-sem })
	}
	wg.Wait()
}

func ParallelMapOrdered[PT, RT any](f func(PT) RT, params []PT, concurrency int) iter.Seq[RT] {
	results := make([]chan RT, len(params))
	for i := range results {
		results[i] = make(chan RT, 1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		sem := make(chan EmptyType, max(concurrency, 1))
		var wg sync.WaitGroup
	work:
		for i, x := range params {
			select {
			case <-ctx.Done():
				break work
			case sem <- Empty:
				wg.Go(func() { results[i] <- f(x); <-sem })
			}
		}
		wg.Wait()
	}()

	return func(yield func(RT) bool) {
		for _, res := range results {
			if !yield(<-res) {
				cancel()
				return
			}
		}
	}
}
