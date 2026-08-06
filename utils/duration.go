package utils

import "time"

type Duration struct{ time.Duration }

func (dur *Duration) UnmarshalText(buf []byte) (e error) {
	var rawDur time.Duration
	if rawDur, e = time.ParseDuration(string(buf)); e == nil {
		*dur = Duration{rawDur}
	}
	return
}

func FormatDuration(msec int64) string {
	return time.Duration(msec * int64(time.Millisecond)).String()
}
