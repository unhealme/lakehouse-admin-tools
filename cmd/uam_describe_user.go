package cmd

import (
	"bufio"
	"os"
	"strings"

	"github.com/pterm/pterm"
	cmd_args "github.com/unhealme/lakehouse-admin-tools/args"
	"github.com/unhealme/lakehouse-admin-tools/internal"
	"github.com/unhealme/lakehouse-admin-tools/internal/uam"
	"github.com/unhealme/lakehouse-admin-tools/utils"
)

const UamDescribeUserVersion = "2026.07.08-0"

func UamDescribeUser(logger *pterm.Logger, args *cmd_args.UamDescribeUserArgs) {
	logger.Debug("using describe user args.", logger.Args(internal.ToArgs(*args)...))

	userInputs := args.Users
	if args.InputFile != "" {
		file, err := os.Open(args.InputFile)
		if err != nil {
			logger.Fatal("unable to read input file.", logger.Args("file", args.InputFile))
		}
		scanner := bufio.NewScanner(file)
		line := 1
		for scanner.Scan() {
			userInputs = append(userInputs, strings.TrimSpace(scanner.Text()))
			line++
		}
		file.Close()
		if scanner.Err() != nil {
			logger.Fatal("unable to read input file.", logger.Args("file", args.InputFile, "line", line))
		}
	}
	if len(userInputs) < 1 {
		logger.Fatal("no user input is specified.")
	}

	var outBuffer *bufio.Writer
	switch args.Format {
	case cmd_args.UamDescribeUsePrintFormatCsv:
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
	case cmd_args.UamDescribeUsePrintFormatDefault:
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
		outBuffer = bufio.NewWriter(outFile)
		defer outBuffer.Flush()
	}

	for i, user := range userInputs {
		entries, err := args.UamClient.DescribeUser(args.BaseDn, user)
		if err != nil {
			logger.Error("unable to describe user.", logger.Args("user", user, "error", err))
			continue
		}
		for _, entry := range entries {
			switch args.Format {
			case cmd_args.UamDescribeUsePrintFormatDefault:
				if i > 0 {
					if _, err := outBuffer.WriteString("\n"); err != nil {
						panic(err)
					}
				}
				uam.PrintDefault(entry, args.GroupBase, outBuffer)
			case cmd_args.UamDescribeUsePrintFormatCsv:
				utils.WriteCsv(uam.EntryCsvSer{Entry: entry, GroupBase: args.GroupBase})
			}
		}
	}
}
