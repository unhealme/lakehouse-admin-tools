package args

import "github.com/unhealme/lakehouse-admin-tools/internal/uam"

type UamListMembersArgs struct {
	Groups []string `arg:"positional,required" placeholder:"GROUP"`
	Unsafe bool     `arg:"-,--unsafe" help:"do not escape GROUP"`

	BaseDn    string         `arg:"-"`
	UamClient *uam.UamClient `arg:"-"`
}
