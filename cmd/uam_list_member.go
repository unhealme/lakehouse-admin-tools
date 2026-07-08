package cmd

import (
	"fmt"
	"slices"
	"strings"

	"github.com/pterm/pterm"
	"github.com/unhealme/lakehouse-admin-tools/internal"
)

const UamListMembersVersion = "2026.07.08-0"

func UamListMembers(logger *pterm.Logger, args *UamListMembersArgs) {
	logger.Debug("using list member args.", logger.Args(internal.ToArgs(*args)...))
	for _, group := range args.Groups {
		entries, err := args.UamClient.ListMembers(args.BaseDn, group)
		if err != nil {
			logger.Error("unable to list member.", logger.Args("group", group, "error", err))
		}
		var members []string
		for _, entry := range entries {
			members = append(members, entry.GetAttributeValue("sAMAccountName"))
		}
		slices.Sort(members)
		fmt.Printf("%s : %s\n", group, strings.Join(members, ","))
	}
}
