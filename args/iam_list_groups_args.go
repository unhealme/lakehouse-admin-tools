package args

import "github.com/unhealme/lakehouse-admin-tools/internal/iam"

type IamListGroupsArgs struct {
	OutputFile string `arg:"-o,--" help:"write result to FILE instead of stdout" placeholder:"FILE"`
	NoHeader   bool   `arg:"-,--no-header" help:"do not print output header"`
	NoProg     bool   `arg:"-,--no-progress" help:"disable progress bar"`

	DomainId  string         `arg:"-"`
	IamClient *iam.IamClient `arg:"-"`
}
