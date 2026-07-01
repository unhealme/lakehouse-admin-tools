package cmd

import (
	"github.com/unhealme/lakehouse-admin-tools/internal"
	"github.com/unhealme/lakehouse-admin-tools/internal/yarn"
)

type YarnAutoKillAppsArgs struct {
	LongerThan internal.Duration `arg:"-,--longer-than,required" placeholder:"DUR" help:"kill yarn applications that running longer than DUR"`
	DryRun     bool              `arg:"-,--dry-run" help:"simulate action without doing anything"`
	NoProg     bool              `arg:"-,--no-progress" help:"disable progress bar"`

	YarnClient *yarn.YarnRMClient `arg:"-"`
}
