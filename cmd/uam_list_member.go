package cmd

import (
	"github.com/pterm/pterm"
	"github.com/unhealme/lakehouse-admin-tools/internal"
)

const UamListMembersVersion = "2026.06.24-1"

func UamListMembers(logger *pterm.Logger, args *UamListMembersArgs) {
	logger.Debug("using list member args.", logger.Args(internal.ToArgs(*args)...))
	for _, group := range args.Groups {
		if err := args.UamClient.ListMembers(group); err != nil {
			logger.Warn("unable to list member.", logger.Args("group", group, "error", err))
		}
	}
}
