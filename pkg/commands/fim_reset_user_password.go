package commands

import (
	"github.com/pterm/pterm"
	"github.com/unhealme/lakehouse-admin-tools/internal"
	"github.com/unhealme/lakehouse-admin-tools/pkg/arguments"
	"github.com/unhealme/lakehouse-admin-tools/pkg/utils"
)

const FimResetUserPasswordVersion = "2026.08.19-0"

func FimResetUserPassword(logger *pterm.Logger, args *arguments.FimResetUserPasswordArgs) {
	logger.Debug("using reset user password args.", logger.Args(internal.ToArgs(*args)...))

	if err := args.FimClient.BasicLogin(args.LoginUser, args.LoginPass); err != nil {
		logger.Fatal("unable to login to FIM.", logger.Args("error", err))
	}

	var prog *utils.ProgressBar
	if !args.NoProg {
		prog, _ = utils.NewProgressBar(pterm.DefaultProgressbar.WithTitle("Resetting password").WithTotal(len(args.Users))).Start()
		defer prog.Stop()
	}
	for _, user := range args.Users {
		if err := args.FimClient.ResetUserPassword(user, args.DefaultPass); err != nil {
			logger.Warn("unable to reset user password.", logger.Args("user", user, "error", err))
		} else {
			logger.Info("user password resetted.", logger.Args("user", user))
		}
		if prog != nil {
			prog.Increment()
		}
	}
}
