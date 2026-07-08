package cmd

import (
	"github.com/unhealme/lakehouse-admin-tools/internal/fim"
	"github.com/unhealme/lakehouse-admin-tools/internal/mrs"
)

type MrsListHetuTenantsArgs struct {
	OutputFile   string `arg:"-o,--" placeholder:"FILE" help:"write result to FILE instead of stdout"`
	FimClusterId int    `arg:"-,--fim-cluster-id" default:"1" placeholder:"NUM"`
	NoHeader     bool   `arg:"-,--no-header" help:"do not print header"`

	MrsClient    *mrs.MrsClient `arg:"-"`
	FimClient    *fim.FimClient `arg:"-"`
	LoginUser    string         `arg:"-"`
	MrsClusterId string         `arg:"-"`
}
