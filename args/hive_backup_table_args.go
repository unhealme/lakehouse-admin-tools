package args

import (
	"errors"
	"strings"

	"github.com/unhealme/lakehouse-admin-tools/internal/hive"
)

type HiveBackupTableArgs struct {
	Tables      []string                 `arg:"positional" placeholder:"[DB.]TABLE"`
	Fixed       bool                     `arg:"-F,--fixed" help:"TABLE is not pattern"`
	Format      HiveBackupTableOutputFmt `arg:"-f,--format" default:"csv" help:"output format" placeholder:"{csv,json}"`
	Concurrency int                      `arg:"-j,--" default:"2" help:"max job concurrency" placeholder:"NUM"`
	InputFile   string                   `arg:"-i,--" help:"read table input from FILE" placeholder:"FILE"`
	OutputFile  string                   `arg:"-o,--" help:"write result to FILE instead of stdout" placeholder:"FILE"`
	NoHeader    bool                     `arg:"-,--no-header" help:"do not print header for csv output format"`
	NoProg      bool                     `arg:"-,--no-progress" help:"disable progress bar"`

	HiveServerClient *hive.HiveServerClient `arg:"-"`
}

type HiveBackupTableOutputFmt int

const (
	HiveBackupTableOutputCsv HiveBackupTableOutputFmt = iota + 1
	HiveBackupTableOutputJson
)

func (f *HiveBackupTableOutputFmt) UnmarshalText(buf []byte) (e error) {
	switch fmt := string(buf); strings.TrimSpace(strings.ToLower(fmt)) {
	case "csv":
		*f = HiveBackupTableOutputCsv
	case "json":
		*f = HiveBackupTableOutputJson
	default:
		return errors.New("invalid output format: " + fmt)
	}
	return
}
