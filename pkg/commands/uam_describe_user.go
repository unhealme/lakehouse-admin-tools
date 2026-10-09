package commands

import (
	"bufio"
	"os"
	"strings"

	"github.com/pterm/pterm"
	"github.com/unhealme/lakehouse-admin-tools/internal"
	"github.com/unhealme/lakehouse-admin-tools/internal/clients/uam"
	"github.com/unhealme/lakehouse-admin-tools/pkg/arguments"
	"github.com/unhealme/lakehouse-admin-tools/pkg/utils"
)

const UamDescribeUserVersion = "2026.10.09-0"

func UamDescribeUser(logger *pterm.Logger, args *arguments.UamDescribeUserArgs) {
	logger.Debug("using describe user args.", logger.Args(internal.ToArgs(*args)...))

	userInputs := args.Users
	if args.InputFile != "" || len(args.Users) < 1 {
		r := os.Stdin
		if args.InputFile != "" && args.InputFile != "-" {
			var err error
			r, err = os.Open(args.InputFile)
			if err != nil {
				logger.Fatal("unable to read input file.", logger.Args("file", args.InputFile))
			}
		}
		re := make(chan error, 1)
		for i := range utils.IterLinesSeq(r, args.InputSep, re) {
			if t := strings.TrimSpace(i); t != "" {
				userInputs = append(userInputs, strings.TrimSpace(i))
			}
		}
		if err := <-re; err != nil {
			logger.Fatal("unable to read input file.", logger.Args("file", args.InputFile, "error", err))
		}
		if r != os.Stdin {
			r.Close()
		}
	}
	if len(userInputs) < 1 {
		logger.Fatal("no user input is specified.")
	}

	var outBuffer *bufio.Writer
	switch args.Format {
	case arguments.UamDescribeUsePrintFormatCsv:
		var headers []string
		if !args.NoHeader {
			headers = []string{
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
			}
		}
		if err := utils.OpenCsvWriter(args.OutputFile, headers); err != nil {
			logger.Fatal("unable to open file to write.", logger.Args("file", args.OutputFile, "error", err))
		}
		defer utils.CloseOutput()
	case arguments.UamDescribeUsePrintFormatDefault:
		outFile := os.Stdout
		if args.OutputFile != "" {
			var err error
			if outFile, err = os.Create(args.OutputFile); err != nil {
				logger.Fatal("unable to open file to write.", logger.Args("file", args.OutputFile, "error", err))
			}
			defer outFile.Close()
		}
		outBuffer = bufio.NewWriter(outFile)
		defer outBuffer.Flush()
	}

	for i, user := range userInputs {
		entries, err := args.UamClient.DescribeUser(args.BaseDn, user, args.Unsafe)
		if err != nil {
			logger.Error("unable to describe user.", logger.Args("user", user, "error", err))
			continue
		}
		for _, entry := range entries {
			switch args.Format {
			case arguments.UamDescribeUsePrintFormatDefault:
				if i > 0 {
					if _, err := outBuffer.WriteString("\n"); err != nil {
						panic(err)
					}
				}
				uam.PrintDefault(entry, args.GroupBase, outBuffer)
			case arguments.UamDescribeUsePrintFormatCsv:
				utils.WriteCsv(uam.EntryCsvSer{Entry: entry, GroupBase: args.GroupBase})
			}
		}
	}
}
