package internal

import (
	"time"
)

type EmptyType struct{}

var Empty EmptyType

type Duration struct{ time.Duration }

func (d *Duration) UnmarshalText(buf []byte) error {
	dur, err := time.ParseDuration(string(buf))
	if err != nil {
		return err
	}
	*d = Duration{dur}
	return nil
}
