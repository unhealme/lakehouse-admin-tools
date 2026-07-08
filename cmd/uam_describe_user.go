package cmd

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"os"
	"strings"

	"github.com/pterm/pterm"
	"github.com/unhealme/lakehouse-admin-tools/internal"
	"github.com/unhealme/lakehouse-admin-tools/internal/uam"
)

const UamDescribeUserVersion = "2026.07.08-0"

func UamDescribeUser(logger *pterm.Logger, args *UamDescribeUserArgs) {
	logger.Debug("using describe user args.", logger.Args(internal.ToArgs(*args)...))
	printFmt := uam.PrintFormatDefault
	switch strings.TrimSpace(strings.ToLower(args.Format)) {
	case "default":
	case "csv":
		printFmt = uam.PrintFormatCSV
	default:
		logger.Fatal(fmt.Sprintf("invalid output format: %s", args.Format))
	}

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
		if scanner.Err() != nil {
			logger.Fatal("unable to read input file.", logger.Args("file", args.InputFile, "line", line))
		}
		file.Close()
	}
	if len(userInputs) < 1 {
		logger.Fatal("no user input is specified.")
	}

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

	var csvWriter *csv.Writer
	if printFmt == uam.PrintFormatCSV {
		csvWriter = csv.NewWriter(outFile)
		if !args.NoHeader {
			csvWriter.Write(
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
		defer csvWriter.Flush()
	}

	for i, user := range userInputs {
		entries, err := args.UamClient.DescribeUser(args.BaseDn, user)
		if err != nil {
			logger.Error("unable to describe user.", logger.Args("user", user, "error", err))
			continue
		}
		for _, entry := range entries {
			switch printFmt {
			case uam.PrintFormatDefault:
				if i > 0 {
					outFile.WriteString("\n")
				}
				uam.PrintDefault(entry, args.GroupBase, outFile)
			case uam.PrintFormatCSV:
				uam.PrintCSV(entry, args.GroupBase, csvWriter)
			}
		}
	}
}
