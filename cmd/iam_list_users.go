package cmd

import (
	"encoding/csv"
	"os"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/pterm/pterm"
	"github.com/unhealme/lakehouse-admin-tools/internal"
)

const IamListUsersVersion = "2026.07.23-0"

func IamListUsers(logger *pterm.Logger, args *IamListUsersArgs) {
	logger.Debug("using iam list users args.", logger.Args(internal.ToArgs(*args)...))

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
	if !args.NoHeader {
		if err := csvWriter.Write(
			[]string{
				"Name",
				"Id",
				"Description",
				"Enabled",
				"AccessMode",
				"Groups",
				"LastLogin",
			}); err != nil {
			panic(err)
		}
	}
	defer csvWriter.Flush()

	users, err := args.IamClient.GetUsers(args.DomainId, true)
	if err != nil {
		logger.Fatal("unable to list IAM users.", logger.Args("error", err))
	}

	results := make([]chan []string, len(users))
	for i := range results {
		results[i] = make(chan []string)
	}
	var wg sync.WaitGroup
	wg.Go(func() {
		for _, r := range results {
			if err := csvWriter.Write(<-r); err != nil {
				panic(err)
			}
		}
	})

	var prog *pterm.ProgressbarPrinter
	if !args.NoProg && args.OutputFile != "" {
		prog, _ = internal.NewProgressBar().WithTitle("Listing users").WithTotal(len(users)).Start()
	}

	sem := make(chan internal.EmptyType, max(args.Concurrency, 1))
	for i, user := range users {
		sem <- internal.Empty
		wg.Go(func() {
			lastLogin, err := args.IamClient.GetUserLastLogin(user.Id)
			if err != nil {
				logger.Warn("unable to get user last login.", logger.Args("user", user.Name, "error", err))
			}
			var lastLoginStr string
			if lastLogin != nil {
				lastLoginStr = lastLogin.Local().Format(time.DateTime)
			}

			groups, err := args.IamClient.GetUserGroups(user.Id)
			if err != nil {
				logger.Warn("unable to get user groups.", logger.Args("user", user.Name, "error", err))
			}
			groupNames := make([]string, len(groups))
			for i, group := range groups {
				groupNames[i] = group.Name
			}
			slices.Sort(groupNames)

			var desc string
			if user.Description != nil {
				desc = *user.Description
			}

			var accessMode string
			if user.AccessMode != nil {
				accessMode = *user.AccessMode
			}
			results[i] <- []string{
				user.Name,
				user.Id,
				desc,
				strconv.FormatBool(user.Enabled),
				accessMode,
				strings.Join(groupNames, ","),
				lastLoginStr,
			}
			if prog != nil {
				prog.Increment()
			}
			<-sem
		})
	}
	wg.Wait()
	if prog != nil {
		prog.Stop()
	}
}
