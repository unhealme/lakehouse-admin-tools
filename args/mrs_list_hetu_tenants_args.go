package args

import (
	"github.com/unhealme/lakehouse-admin-tools/internal/fim"
	"github.com/unhealme/lakehouse-admin-tools/internal/mrs"
)

type MrsListHetuTenantsArgs struct {
	OutputFile   string `arg:"-o,--" placeholder:"FILE" help:"write result to FILE instead of stdout"`
	FimClusterId int    `arg:"-,--fim-cluster-id" default:"1" placeholder:"NUM"`
	FimAddress   string `arg:"-,--fim-url,required" placeholder:"FIM_ADDRESS"`
	NoHeader     bool   `arg:"-,--no-header" help:"do not print header"`
	LoginUser    string `arg:"-,--user,env:FIM_USER" placeholder:"FIM_USER"`

	MrsClient    *mrs.MrsClient `arg:"-"`
	MrsClusterId string         `arg:"-"`
	FimClient    *fim.FimClient `arg:"-"`
}
