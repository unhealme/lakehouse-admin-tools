package commands

import (
	json "encoding/json/v2"
	"strconv"
	"time"

	"github.com/unhealme/lakehouse-admin-tools/internal"
	"github.com/unhealme/lakehouse-admin-tools/internal/clients/yarn"
	"github.com/unhealme/lakehouse-admin-tools/internal/logger"
	"github.com/unhealme/lakehouse-admin-tools/pkg/arguments"
	"github.com/unhealme/lakehouse-admin-tools/pkg/utils"
)

const YarnListAppsVersion = "2026.09.17-0"

func YarnListApps(args *arguments.YarnListAppsArgs) {
	logger.Debug("using list apps args.", logger.Args(internal.ToArgs(*args)...))

	outputHeaders := []string{
		"Id",
		"Name",
		"Type",
		"QueueUser",
		"Queue",
		"StartTime",
		"FinishTime",
		"FinalState",
		"Memory",
		"CPU",
	}
	switch args.Format {
	case arguments.YarnListAppsOutputCsv:
		var headers []string
		if !args.NoHeader {
			headers = outputHeaders
		}
		if err := utils.OpenCsvWriter(args.OutputFile, headers); err != nil {
			logger.Fatal("unable to open file to write.", logger.Args("file", args.OutputFile, "error", err))
		}
	case arguments.YarnListAppsOutputJson:
		if err := utils.OpenJsonWriter(args.OutputFile); err != nil {
			logger.Fatal("unable to open file to write.", logger.Args("file", args.OutputFile, "error", err))
		}
	case arguments.YarnListAppsOutputTable:
		if err := utils.OpenTableWriter(args.OutputFile, outputHeaders); err != nil {
			logger.Fatal("unable to open file to write.", logger.Args("file", args.OutputFile, "error", err))
		}
	}
	defer utils.CloseOutput()

	apps, err := args.YarnClient.Applications([]yarn.ApplicationState(args.States), args.User, args.Queue, args.Limit)
	if err != nil {
		logger.Fatal("unable to get yarn applications.", logger.Args("error", err))
	}
	logger.Info("yarn applications fetched.", logger.Args("count", len(apps.Apps.App)))

	for _, app := range apps.Apps.App {
		utils.WriteOutput(yarnListAppsResult(app))
	}
}

type yarnListAppsResult yarn.Application

func (a yarnListAppsResult) parseFinishedTime() string {
	if a.FinishedTime < 1 {
		return ""
	}
	return time.Unix(0, a.FinishedTime*1e6).Local().Format(time.DateTime)
}

func (a yarnListAppsResult) parseStartedTime() string {
	if a.StartedTime < 1 {
		return ""
	}
	return time.Unix(0, a.StartedTime*1e6).Local().Format(time.DateTime)
}

func (a yarnListAppsResult) SerCsv() []string {
	return []string{
		a.Id,
		a.Name,
		a.ApplicationType,
		a.QueueUser,
		a.Queue,
		a.parseStartedTime(),
		a.parseFinishedTime(),
		a.FinalStatus,
		strconv.FormatInt(a.MemorySeconds, 10),
		strconv.FormatInt(a.VcoreSeconds, 10),
	}
}

func (a yarnListAppsResult) SerJson() (v []byte) {
	v, _ = json.Marshal(a)
	return
}

func (a yarnListAppsResult) SerTable() (v []string) {
	v = a.SerCsv()
	for n, i := range v {
		if len(i) > 50 {
			v[n] = i[:51] + "..."
		}
	}
	return
}
