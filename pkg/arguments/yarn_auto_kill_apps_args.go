package arguments

import (
	"github.com/unhealme/lakehouse-admin-tools/internal/clients/yarn"
	"github.com/unhealme/lakehouse-admin-tools/pkg/utils"
)

type YarnAutoKillAppsArgs struct {
	LongerThan utils.Duration `arg:"-,--longer-than,required" placeholder:"DUR" help:"kill yarn applications that running longer than DUR"`
	DryRun     bool           `arg:"-,--dry-run" help:"simulate action without doing anything"`
	NoProg     bool           `arg:"-,--no-progress" help:"disable progress bar"`

	YarnClient *yarn.YarnRMClient `arg:"-"`
}
