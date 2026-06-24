package uam

import (
	"github.com/unhealme/lakehouse-admin-tools/cmd"
	"github.com/unhealme/lakehouse-admin-tools/internal"
)

func ListMembers(logger *internal.Logger, args *cmd.UamListMembersArgs) {
	logger.Debug("using list member args.", logger.Args(internal.ToArgs(*args)...))
	for _, group := range args.Groups {
		if err := args.UamClient.ListMembers(group); err != nil {
			logger.Warn("unable to list member.", logger.Args("group", group, "error", err))
		}
	}
}
