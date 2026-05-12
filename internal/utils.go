package internal

import (
	"errors"
	"fmt"
	"iter"
	"os"
	"reflect"
	"strings"
	"sync"
	"time"

	timefmt "github.com/itchyny/timefmt-go"
)

func GetEnv(k, def string) string {
	if v, e := os.LookupEnv(k); e {
		return v
	}
	return def
}

func ParallelMap[T any](task func(T), pool iter.Seq[T], concurrency int) {
	sem := make(chan EmptyType, max(concurrency, 1))
	var wg sync.WaitGroup
	for i := range pool {
		sem <- Empty
		wg.Go(func() { task(i); <-sem })
	}
	wg.Wait()
}

func ParseStrftime(rawDt, format string, target *time.Time) error {
	if target == nil {
		return errors.New("required non nil pointer")
	}
	dt, err := timefmt.Parse(rawDt, format)
	if err != nil {
		return fmt.Errorf("parsing date %q with format %q: %w", rawDt, format, err)
	}
	*target = dt
	return nil
}

func SliceToSet[T comparable](s []T) map[T]EmptyType {
	m := make(map[T]EmptyType)
	for _, i := range s {
		m[i] = Empty
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

func FormatDuration(msec int64) string {
	return time.Duration(msec * int64(time.Millisecond)).String()
}
