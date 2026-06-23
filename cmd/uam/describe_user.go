package uam

import (
	"fmt"

	"github.com/unhealme/lakehouse-admin-tools/cmd"
	"github.com/unhealme/lakehouse-admin-tools/internal"
)

func DescribeUser(logger *internal.Logger, args *cmd.UamDescribeUserArgs) {
	logger.Debug("using describe user args.", logger.Args(internal.ToArgs(*args)...))
	for i, user := range args.Users {
		if i > 0 {
			fmt.Println()
		}
		if err := args.UamClient.DescribeUser(user); err != nil {
			logger.Warn("unable to describe user.", logger.Args("user", user, "error", err))
		}
	}
}
