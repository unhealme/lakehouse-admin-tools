package arguments

import (
	"github.com/unhealme/lakehouse-admin-tools/internal/clients/fim"
	"github.com/unhealme/lakehouse-admin-tools/internal/clients/mrs"
)

type MrsDumpHetuClustersArgs struct {
	OutputFile   string `arg:"-o,--" placeholder:"FILE" help:"write result to FILE instead of stdout"`
	FimClusterId int    `arg:"-,--fim-cluster-id" default:"1" placeholder:"NUM"`
	FimAddress   string `arg:"-,--fim-url,required" placeholder:"FIM_ADDRESS"`
	LoginUser    string `arg:"-,--user,env:FIM_USER" placeholder:"FIM_USER"`

	MrsClient    *mrs.MrsClient `arg:"-"`
	MrsClusterId string         `arg:"-"`
	FimClient    *fim.FimClient `arg:"-"`
}
