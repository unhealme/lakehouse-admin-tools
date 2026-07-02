package cmd

import (
	"encoding/csv"
	"fmt"
	"os"
	"strings"

	"github.com/pterm/pterm"
	"github.com/unhealme/lakehouse-admin-tools/internal"
	"github.com/unhealme/lakehouse-admin-tools/internal/uam"
)

const UamDescribeUserVersion = "2026.07.02-0"

func UamDescribeUser(logger *pterm.Logger, args *UamDescribeUserArgs) {
	logger.Debug("using describe user args.", logger.Args(internal.ToArgs(*args)...))
	var writer *csv.Writer
	printFmt := uam.PrintFormatDefault
	switch strings.TrimSpace(strings.ToLower(args.Format)) {
	case "default":
	case "csv":
		printFmt = uam.PrintFormatCSV
		writer = csv.NewWriter(os.Stdout)
		if !args.NoHeader {
			writer.Write(
				[]string{
					"name",
					"username",
					"mail",
					"department",
					"directorate",
					"divisionGroup",
					"division",
					"group",
					"distinguishedName",
					"badPwdCount",
					"badPasswordTime",
					"lockoutTime",
					"pwdLastSet",
					"lastLogon",
				})
		}
		defer writer.Flush()
	default:
		logger.Fatal(fmt.Sprintf("invalid output format: %s", args.Format))
	}
	for i, user := range args.Users {
		if i > 0 && printFmt == uam.PrintFormatDefault {
			fmt.Println()
		}
		if err := args.UamClient.DescribeUser(user, printFmt, writer); err != nil {
			logger.Warn("unable to describe user.", logger.Args("user", user, "error", err))
		}

	}
}
