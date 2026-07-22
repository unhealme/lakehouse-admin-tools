package cmd

import "github.com/unhealme/lakehouse-admin-tools/internal/obs"

type ObsAnalyzeArgs struct {
	Paths       []string `arg:"positional,required" placeholder:"PATH"`
	Fixed       bool     `arg:"-F,--" help:"PATH is fixed string"`
	Concurrency int      `arg:"-j,--" default:"4" help:"max job concurrency" placeholder:"NUM"`
	Summarize   bool     `arg:"-s,--summarize" help:"show total statistics for all input paths"`
	CsvOut      string   `arg:"-,--write-csv" help:"write csv format output to FILE" placeholder:"FILE"`
	JsonOut     string   `arg:"-,--write-json" help:"write json format output to FILE" placeholder:"FILE"`
	NoProg      bool     `arg:"-,--no-progress" help:"disable progress bar"`

	ObsClient *obs.ObsClient `arg:"-"`
}
