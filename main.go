package main

import (
	"os"
	"strings"

	arg "github.com/alexflint/go-arg"
	"github.com/pterm/pterm"
	"github.com/unhealme/lakehouse-admin-tools/cmd"
	"github.com/unhealme/lakehouse-admin-tools/config"
	"github.com/unhealme/lakehouse-admin-tools/internal"
	"github.com/unhealme/lakehouse-admin-tools/internal/dataarts"
	"github.com/unhealme/lakehouse-admin-tools/internal/fim"
	"github.com/unhealme/lakehouse-admin-tools/internal/iam"
	"github.com/unhealme/lakehouse-admin-tools/internal/mrs"
	"github.com/unhealme/lakehouse-admin-tools/internal/obs"
	"github.com/unhealme/lakehouse-admin-tools/internal/uam"
	"github.com/unhealme/lakehouse-admin-tools/internal/yarn"
)

var logger = pterm.DefaultLogger.WithLevel(pterm.LogLevelInfo).WithWriter(os.Stderr)

func main() {
	var args config.Arguments
	arg.MustParse(&args)
	if args.Verbose {
		logger = logger.WithLevel(pterm.LogLevelDebug)
	}
	if args.NoColor {
		pterm.DisableColor()
	}
	logger.Debug("parsed arguments.", logger.Args(internal.ToArgs(args)...))
	cfg := config.GetConfig(logger, args.ConfigFile)
	arg.MustParse(cfg)
	logger.Debug("current config.", logger.Args(internal.ToArgs(*cfg)...))

	switch {
	case args.DataArts != nil:
		logger.Debug("creating DataArts Studio client.")
		dasClient, err := dataarts.NewClient(cfg.AccessKey, cfg.SecretKey, cfg.SessionToken, cfg.Region)
		if err != nil {
			logger.Fatal("unable to create DataArts Studio client.", logger.Args("error", err))
		}

		logger.Debug("creating IAM client.")
		iamClient, err := iam.NewClient(cfg.AccessKey, cfg.SecretKey, cfg.SessionToken, cfg.Region)
		if err != nil {
			logger.Fatal("unable to create IAM client.", logger.Args("error", err))
		}

		switch {
		case args.DataArts.CreateHetuConnection != nil:
			subArgs := args.DataArts.CreateHetuConnection
			subArgs.DomainId = cfg.DomainId
			subArgs.InstanceId = cfg.DataArts.InstanceId
			subArgs.DataArtsClient = dasClient
			subArgs.IamClient = iamClient
			subArgs.HetuConfig = cfg.DataArts.HetuConfig
			if subArgs.AgentId == "" {
				subArgs.AgentId = cfg.DataArts.Agent.Id
			}
			if subArgs.AgentName == "" {
				subArgs.AgentName = cfg.DataArts.Agent.Name
			}

			cmd.DataArtsCreateHetuConnection(logger, subArgs)
		case args.DataArts.UpdateHetuConnection != nil:
			subArgs := args.DataArts.UpdateHetuConnection
			subArgs.DomainId = cfg.DomainId
			subArgs.InstanceId = cfg.DataArts.InstanceId
			subArgs.DataArtsClient = dasClient
			subArgs.HetuConfig = cfg.DataArts.HetuConfig

			cmd.DataArtsUpdateHetuConnection(logger, subArgs)
		}
	case args.Mrs != nil:
		mrsClient, err := mrs.NewClient(cfg.AccessKey, cfg.SecretKey, cfg.SessionToken, cfg.Region)
		if err != nil {
			logger.Fatal("unable to create MRS client.", logger.Args("error", err))
		}

		fimClient, err := fim.NewClient(cfg.Mrs.FimAddress)
		if err != nil {
			logger.Fatal("unable to create FIM client.", logger.Args("error", err))
		}
		defer fimClient.Close()

		switch {
		case args.Mrs.ListHetuTenants != nil:
			subArgs := args.Mrs.ListHetuTenants
			subArgs.MrsClient = mrsClient
			subArgs.FimClient = fimClient
			subArgs.LoginUser = cfg.Mrs.LoginUser
			subArgs.MrsClusterId = cfg.Mrs.ClusterId

			cmd.MrsListHetuTenants(logger, subArgs)
		}
	case args.Obs != nil:
		obsClient, err := obs.NewClient(cfg.Obs.Endpoint, cfg.AccessKey, cfg.SecretKey, cfg.SessionToken)
		if err != nil {
			logger.Fatal("unable to create OBS client.", logger.Args("error", err))
		}
		defer obsClient.Close()

		switch {
		case args.Obs.Analyze != nil:
			subArgs := args.Obs.Analyze
			subArgs.ObsClient = obsClient

			cmd.ObsAnalyze(logger, subArgs)
		case args.Obs.BatchRename != nil:
			subArgs := args.Obs.BatchRename
			subArgs.ObsClient = obsClient
			if !strings.HasSuffix(subArgs.Path, "/") {
				subArgs.Path += "/"
			}

			cmd.ObsBatchRename(logger, subArgs)
		case args.Obs.BatchSetStorageClass != nil:
			subArgs := args.Obs.BatchSetStorageClass
			subArgs.ObsClient = obsClient

			cmd.ObsBatchSetStorageClass(logger, subArgs)
		}
	case args.Ps != nil:
		switch {
		case args.Ps.AutoKill != nil:
			cmd.PsAutoKill(logger, args.Ps.AutoKill)
		}
	case args.Uam != nil:
		uamClient, err := uam.NewClient(
			logger, cfg.Uam.Url, cfg.Uam.User, cfg.Uam.Password,
			cfg.Uam.MailDomain, cfg.Uam.Realm,
		)
		if err != nil {
			logger.Fatal("unable to create UAM client.", logger.Args("error", err))
		}
		defer uamClient.Close()

		switch {
		case args.Uam.DescribeUser != nil:
			subArgs := args.Uam.DescribeUser
			subArgs.BaseDn = cfg.Uam.BaseDN
			subArgs.GroupBase = cfg.Uam.GroupBase
			subArgs.UamClient = uamClient

			cmd.UamDescribeUser(logger, subArgs)
		case args.Uam.ListMembers != nil:
			subArgs := args.Uam.ListMembers
			subArgs.BaseDn = cfg.Uam.BaseDN
			subArgs.UamClient = uamClient

			cmd.UamListMembers(logger, subArgs)
		}
	case args.Yarn != nil:
		yarnClient, err := yarn.NewClient(cfg.Yarn.RMAddress)
		if err != nil {
			logger.Fatal("unable to create YARN client.", logger.Args("error", err))
		}
		defer yarnClient.Close()

		switch {
		case args.Yarn.AutoKillApps != nil:
			subArgs := args.Yarn.AutoKillApps
			subArgs.YarnClient = yarnClient

			cmd.YarnAutoKillApps(logger, subArgs)
		}
	}
}
