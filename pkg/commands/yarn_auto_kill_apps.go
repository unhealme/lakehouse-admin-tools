package commands

import (
	"github.com/pterm/pterm"
	"github.com/unhealme/lakehouse-admin-tools/internal"
	"github.com/unhealme/lakehouse-admin-tools/internal/clients/yarn"
	"github.com/unhealme/lakehouse-admin-tools/internal/logger"
	"github.com/unhealme/lakehouse-admin-tools/pkg/arguments"
	"github.com/unhealme/lakehouse-admin-tools/pkg/utils"
)

const YarnAutoKillAppsVersion = "2026.10.05-0"

func YarnAutoKillApps(args *arguments.YarnAutoKillAppsArgs) {
	logger.Debug("using auto kill apps args.", logger.Args(internal.ToArgs(*args)...))
	apps, err := args.YarnClient.Applications([]yarn.ApplicationState{yarn.RUNNING}, "", "", 0)
	if err != nil {
		logger.Fatal("unable to get yarn applications.", logger.Args("error", err))
	}

	logger.Info("yarn applications fetched.", logger.Args("count", len(apps.Apps.App)))
	var appToKill []yarn.Application
	var total int
	for _, app := range apps.Apps.App {
		if app.ElapsedTime >= args.LongerThan.Milliseconds() {
			appToKill = append(appToKill, app)
			total += 1
		}
	}

	if total > 0 {
		logger.Info("yarn applications filtered.", logger.Args("app to kill", total))
		var prog *utils.ProgressBar
		if !args.NoProg {
			prog, _ = utils.NewProgressBar(pterm.DefaultProgressbar.WithTitle("Killing yarn applications").WithTotal(total)).Start()
			defer prog.Stop()
		}
		for _, app := range appToKill {
			var err error
			if !args.DryRun {
				err = args.YarnClient.KillApplication(app)
			}
			if err != nil {
				logger.Error("unable to kill yarn application.", logger.Args("id", app.Id, "error", err))
			} else {
				logger.Info(
					"yarn application killed.",
					logger.Args(
						"id", app.Id,
						"name", app.Name,
						"queue", app.Queue,
						"engine", app.ApplicationType,
						"elapsed", utils.FormatDuration(app.ElapsedTime),
					),
				)
			}
			if prog != nil {
				prog.Increment()
			}
		}
	}
	logger.Info("auto kill yarn applications done.")
}
