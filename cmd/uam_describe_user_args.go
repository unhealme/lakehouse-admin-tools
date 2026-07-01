package cmd

import "github.com/unhealme/lakehouse-admin-tools/internal/uam"

type UamDescribeUserArgs struct {
	Users    []string `arg:"positional,required" placeholder:"USER"`
	Format   string   `arg:"-f,--format" default:"default" help:"output format" placeholder:"{default,csv}"`
	NoHeader bool     `arg:"-,--no-header" help:"do not print header for csv output format"`

	UamClient *uam.UamClient `arg:"-"`
}
