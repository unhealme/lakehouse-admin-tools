package internal

import (
	"errors"
	"fmt"
	"iter"
	"net/url"
	"os"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	timefmt "github.com/itchyny/timefmt-go"
	"github.com/shirou/gopsutil/v4/process"
)

func BuildUrlEncodedPayload(payload map[string]string) string {
	var array []string
	for k, v := range payload {
		array = append(array, fmt.Sprintf("%s=%s", url.QueryEscape(k), url.QueryEscape(v)))
	}
	return strings.Join(array, "&")
}

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

var sizeUnits = [...]string{
	"K", "M", "G", "T", "P", "E", "Z", "Y", "R",
}

func FormatSize(si int64) string {
	if si < 1024 {
		return strconv.FormatInt(si, 10)
	}

	sf := float64(si)
	for _, u := range sizeUnits {
		if sf /= 1024.0; sf < 1024.0 {
			return fmt.Sprintf("%3.1f%s", sf, u)
		}
	}
	return fmt.Sprintf("%.1fQ", sf)
}

func ParseSize(si string) (int64, error) {
	sizeUnit := strings.ToUpper(si[len(si)-1:])
	if _, err := strconv.ParseFloat(sizeUnit, 64); err == nil {
		return strconv.ParseInt(si, 10, 64)
	}
	size, err := strconv.ParseFloat(si[:len(si)-1], 64)
	if err != nil {
		return 0, err
	}
	mul := 1024.0
	for _, u := range sizeUnits {
		if sizeUnit == u {
			return int64(size * mul), nil
		}
		mul *= 1024.0
	}
	if sizeUnit == "Q" {
		return int64(size * mul), nil
	}
	return 0, fmt.Errorf("unable to parse size %s to int", si)
}

func SoftKill(process *process.Process) error {
	if running, _ := process.IsRunning(); running {
		if err := process.Terminate(); err != nil {
			return err
		}
	} else {
		return nil
	}
	for i := 1; i <= 15; i++ {
		if running, _ := process.IsRunning(); !running {
			return nil
		}
		time.Sleep(1 * time.Second)
	}
	if running, _ := process.IsRunning(); running {
		if err := process.Kill(); err != nil {
			return err
		}
	}
	return nil
}
