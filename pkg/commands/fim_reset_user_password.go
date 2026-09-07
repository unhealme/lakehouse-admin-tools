package commands

import (
	"context"

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

	var prog *pterm.ProgressbarPrinter
	if !args.NoProg {
		ctx, done := context.WithCancel(context.Background())
		prog, _ = utils.NewProgressBar(ctx).WithTitle("Resetting password").WithTotal(len(args.Users)).Start()
		defer done()
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
