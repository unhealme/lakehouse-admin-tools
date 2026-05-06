package internal

import (
	"fmt"
	"iter"
	"reflect"
	"strings"
	"sync"
	"time"

	timefmt "github.com/itchyny/timefmt-go"
)

type emptyType struct{}

var empty emptyType

func ParallelMap[T any](task func(T), pool iter.Seq[T], concurrency int) {
	sem := make(chan emptyType, max(concurrency, 1))
	var wg sync.WaitGroup
	for i := range pool {
		sem <- empty
		wg.Go(func() { task(i); <-sem })
	}
	wg.Wait()
}

func ParseStrftime(rawDt, format string, target *time.Time) error {
	dt, err := timefmt.Parse(rawDt, format)
	if err != nil {
		return fmt.Errorf("parsing date %q with format %q: %w", rawDt, format, err)
	}
	*target = dt
	return nil
}

func SliceToSet[T comparable](s []T) map[T]emptyType {
	m := make(map[T]emptyType)
	for _, i := range s {
		m[i] = empty
	}
	return m
}

func ToArgs(a any) []any {
	var (
		args []any
		v    = reflect.ValueOf(a)
	)
	for _, f := range reflect.VisibleFields(reflect.TypeOf(a)) {
		if !strings.HasPrefix(f.Name, "_") {
			args = append(args, f.Name)
			args = append(args, fmt.Sprintf("%#v", v.FieldByName(f.Name)))
		}
	}
	return args
}
