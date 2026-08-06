package cmd

import (
	"encoding/csv"
	"os"
	"time"

	"github.com/pterm/pterm"
	"github.com/unhealme/lakehouse-admin-tools/args"
	"github.com/unhealme/lakehouse-admin-tools/internal"
)

const IamListGroupsVersion = "2026.08.05-0"

func IamListGroups(logger *pterm.Logger, args *args.IamListGroupsArgs) {
	logger.Debug("using iam list groups args.", logger.Args(internal.ToArgs(*args)...))

	outFile := os.Stdout
	if args.OutputFile != "" {
		var err error
		if outFile, err = os.OpenFile(
			args.OutputFile,
			os.O_WRONLY|os.O_CREATE|os.O_TRUNC,
			0o644,
		); err != nil {
			logger.Fatal("unable to open file to write.", logger.Args("file", args.OutputFile, "error", err))
		}
		defer outFile.Close()
	}
	csvWriter := csv.NewWriter(outFile)
	defer csvWriter.Flush()

	groups, err := args.IamClient.GetGroups(args.DomainId, true)
	if err != nil {
		logger.Fatal("unable to list IAM groups.", logger.Args("error", err))
	}

	if !args.NoHeader {
		if err := csvWriter.Write(
			[]string{
				"Name",
				"Id",
				"Description",
				"CreateTime",
			}); err != nil {
			panic(err)
		}
	}

	for _, group := range groups {
		if err := csvWriter.Write([]string{
			group.Name,
			group.Id,
			group.Description,
			time.UnixMilli(group.CreateTime).String(),
		}); err != nil {
			panic(err)
		}
	}
}
