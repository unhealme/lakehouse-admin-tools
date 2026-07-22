package cmd

import "github.com/unhealme/lakehouse-admin-tools/internal/dataarts"

type DataArtsUpdateHetuConnectionArgs struct {
	UserPairs []string `arg:"positional,required" placeholder:"USER:WORKSPACE"`
	Tenant    string   `arg:"-,--tenant" help:"tenant name" placeholder:"[DEV:]PROD"`

	DomainId       string                   `arg:"-"`
	InstanceId     string                   `arg:"-"`
	DataArtsClient *dataarts.DataArtsClient `arg:"-"`
	HetuConfig     dataarts.DwConfig        `arg:"-"`
}
