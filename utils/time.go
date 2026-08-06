package utils

import (
	"fmt"
	"time"

	"github.com/itchyny/timefmt-go"
)

func ParseStrftime(rawDt, format string) (t *time.Time, e error) {
	if dt, err := timefmt.Parse(rawDt, format); err != nil {
		e = fmt.Errorf("parsing date %q with format %q: %w", rawDt, format, err)
	} else {
		t = &dt
	}
	return
}
