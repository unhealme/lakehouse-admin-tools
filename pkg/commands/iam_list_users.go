package commands

import (
	"encoding/csv"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"

	model "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/iam/v3/model"
	"github.com/pterm/pterm"
	"github.com/unhealme/lakehouse-admin-tools/internal"
	"github.com/unhealme/lakehouse-admin-tools/internal/logger"
	"github.com/unhealme/lakehouse-admin-tools/pkg/arguments"
	"github.com/unhealme/lakehouse-admin-tools/pkg/utils"
)

const IamListUsersVersion = "2026.09.05-0"

func IamListUsers(args *arguments.IamListUsersArgs) {
	logger.Debug("using iam list users args.", logger.Args(internal.ToArgs(*args)...))

	outFile := os.Stdout
	if args.OutputFile != "" {
		var err error
		if outFile, err = os.Create(args.OutputFile); err != nil {
			logger.Fatal("unable to open file to write.", logger.Args("file", args.OutputFile, "error", err))
		}
		defer outFile.Close()
	}
	csvWriter := csv.NewWriter(outFile)
	defer csvWriter.Flush()

	users, err := args.IamClient.GetUsers(args.DomainId, true)
	if err != nil {
		logger.Fatal("unable to list IAM users.", logger.Args("error", err))
	}

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
			},
		); err != nil {
			panic(err)
		}
	}

	var prog *utils.ProgressBar
	if !args.NoProg && args.OutputFile != "" {
		prog, _ = utils.NewProgressBar(pterm.DefaultProgressbar.WithTitle("Listing users").WithTotal(len(users))).Start()
		defer prog.Stop()
	}

	serializeIamUser := func(user *model.KeystoneListUsersResult) []string {
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

		if prog != nil {
			prog.Increment()
		}
		return []string{
			user.Name,
			user.Id,
			desc,
			strconv.FormatBool(user.Enabled),
			accessMode,
			strings.Join(groupNames, ","),
			lastLoginStr,
		}
	}
	for result := range utils.NewSlot(max(args.Concurrency, 1)).MapValue(serializeIamUser, users, false) {
		if err := csvWriter.Write(result); err != nil {
			panic(err)
		}
	}
}
