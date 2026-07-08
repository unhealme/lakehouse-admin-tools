package cmd

import "github.com/unhealme/lakehouse-admin-tools/internal/uam"

type UamDescribeUserArgs struct {
	Users      []string `arg:"positional" placeholder:"USER"`
	InputFile  string   `arg:"-i,--" help:"read user input from FILE" placeholder:"FILE"`
	OutputFile string   `arg:"-o,--" help:"write result to FILE instead of stdout" placeholder:"FILE"`
	Format     string   `arg:"-f,--format" default:"default" help:"output format" placeholder:"{default,csv}"`
	NoHeader   bool     `arg:"-,--no-header" help:"do not print header for csv output format"`

	BaseDn    string         `arg:"-"`
	GroupBase string         `arg:"-"`
	UamClient *uam.UamClient `arg:"-"`
}
