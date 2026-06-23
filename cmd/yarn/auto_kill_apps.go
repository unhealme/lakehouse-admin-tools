package yarn

import (
	"errors"

	"github.com/pterm/pterm"
	"github.com/unhealme/lakehouse-admin-tools/cmd"
	"github.com/unhealme/lakehouse-admin-tools/internal"
	"github.com/unhealme/lakehouse-admin-tools/internal/yarn"
)

func AutoKillApps(logger *internal.Logger, args *cmd.YarnAutoKillAppsArgs) {
	logger.Debug("using auto kill apps args.", logger.Args(internal.ToArgs(*args)...))
	apps, err := args.YarnClient.Applications(logger, []yarn.ApplicationState{yarn.RUNNING}, "", "", 0)
	if err != nil {
		if err, match := errors.AsType[internal.HttpNotOk](err); match {
			args := []any{"status", err.Status, "headers", err.Header}
			if err.Err != nil {
				args = append(args, "error", err.Err)
			}
			if err.Body != nil {
				args = append(args, "body", string(err.Body))
			}
			logger.Fatal("get yarn applications return not ok.", logger.Args(args...))
		}
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
		var prog *pterm.ProgressbarPrinter
		if !args.NoProg {
			prog, _ = internal.NewProgressBar().WithTitle("Killing yarn applications").WithTotal(total).Start()
			defer prog.Stop()
		}
		for _, app := range appToKill {
			ok := true
			if !args.DryRun {
				ok = args.YarnClient.KillApplication(logger, app)
			}
			if ok {
				logger.Info("yarn application killed.",
					logger.Args(
						"id", app.Id,
						"name", app.Name,
						"queue", app.Queue,
						"engine", app.ApplicationType,
						"elapsed", internal.FormatDuration(app.ElapsedTime),
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
