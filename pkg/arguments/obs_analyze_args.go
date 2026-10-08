package arguments

import "github.com/unhealme/lakehouse-admin-tools/internal/clients/obs"

type ObsAnalyzeArgs struct {
	Paths       []string `arg:"positional" placeholder:"PATH"`
	Fixed       bool     `arg:"-F,--" help:"PATH is fixed string"`
	InputFile   string   `arg:"-i,--" help:"read PATH inputs from FILE" placeholder:"FILE"`
	Concurrency int      `arg:"-j,--" default:"4" help:"max job concurrency" placeholder:"NUM"`
	InputSep    string   `arg:"-d,--delimiter" help:"delimiter for input PATH in FILE (default to newline)"`
	Summarize   bool     `arg:"-s,--summarize" help:"show total statistics for all input paths"`
	MinChunks   int      `arg:"-,--chunks" default:"1" help:"split PATH into minimum of NUM chunks" placeholder:"NUM"`
	CsvOut      string   `arg:"-,--write-csv" help:"write csv format output to FILE" placeholder:"FILE"`
	JsonOut     string   `arg:"-,--write-json" help:"write json format output to FILE" placeholder:"FILE"`
	NoProg      bool     `arg:"-,--no-progress" help:"disable progress bar"`

	ObsClient *obs.ObsClient `arg:"-"`
}
