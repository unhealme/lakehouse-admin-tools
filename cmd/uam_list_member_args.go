package cmd

import "github.com/unhealme/lakehouse-admin-tools/internal/uam"

type UamListMembersArgs struct {
	Groups []string `arg:"positional,required" placeholder:"GROUP"`

	UamClient *uam.UamClient `arg:"-"`
}
