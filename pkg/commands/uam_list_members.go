package commands

import (
	"fmt"
	"slices"
	"strings"

	"github.com/unhealme/lakehouse-admin-tools/internal"
	"github.com/unhealme/lakehouse-admin-tools/internal/logger"
	"github.com/unhealme/lakehouse-admin-tools/pkg/arguments"
)

const UamListMembersVersion = "2026.08.19-0"

func UamListMembers(args *arguments.UamListMembersArgs) {
	logger.Debug("using list member args.", logger.Args(internal.ToArgs(*args)...))
	for _, group := range args.Groups {
		groupInfos, err := args.UamClient.ListMembers(args.BaseDn, group, args.Unsafe)
		if err != nil {
			logger.Error("unable to list member.", logger.Args("group", group, "error", err))
		}
		for _, groupInfo := range groupInfos {
			var members []string
			for _, entry := range groupInfo.Members {
				members = append(members, entry.GetAttributeValue("sAMAccountName"))
			}
			slices.Sort(members)
			fmt.Printf("%s : %s\n", groupInfo.Group.GetAttributeValue("cn"), strings.Join(members, ","))
		}
	}
}
