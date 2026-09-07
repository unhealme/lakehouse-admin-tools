package main

import (
	"os"
	"strings"

	arg "github.com/alexflint/go-arg"
	"github.com/pterm/pterm"
	"github.com/unhealme/lakehouse-admin-tools/internal"
	"github.com/unhealme/lakehouse-admin-tools/internal/clients/dataarts"
	"github.com/unhealme/lakehouse-admin-tools/internal/clients/fim"
	"github.com/unhealme/lakehouse-admin-tools/internal/clients/hive"
	"github.com/unhealme/lakehouse-admin-tools/internal/clients/iam"
	"github.com/unhealme/lakehouse-admin-tools/internal/clients/mrs"
	"github.com/unhealme/lakehouse-admin-tools/internal/clients/obs"
	"github.com/unhealme/lakehouse-admin-tools/internal/clients/uam"
	"github.com/unhealme/lakehouse-admin-tools/internal/clients/yarn"
	"github.com/unhealme/lakehouse-admin-tools/pkg/commands"
	"github.com/unhealme/lakehouse-admin-tools/pkg/config"
)

var logger = pterm.DefaultLogger.WithLevel(pterm.LogLevelInfo).WithWriter(os.Stderr).WithMaxWidth(200)

func parseArgs() (*config.Arguments, *config.Arguments) {
	var args config.Arguments
	arg.MustParse(&args)
	logger.Debug("parsed arguments.", logger.Args(internal.ToArgs(args)...))

	if args.Verbose {
		logger = logger.WithLevel(pterm.LogLevelDebug)
	}

	if args.NoColor {
		pterm.DisableColor()
	}

	cfg := config.GetConfig(logger, args.ConfigFile)
	arg.MustParse(cfg)
	logger.Debug("current config.", logger.Args(internal.ToArgs(*cfg)...))

	return &args, cfg
}

func main() {
	args, cfg := parseArgs()
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

			commands.DataArtsCreateHetuConnection(logger, subArgs)
		case args.DataArts.UpdateHetuConnection != nil:
			subArgs := args.DataArts.UpdateHetuConnection
			subArgs.DomainId = cfg.DomainId
			subArgs.InstanceId = cfg.DataArts.InstanceId
			subArgs.DataArtsClient = dasClient
			subArgs.HetuConfig = cfg.DataArts.HetuConfig

			commands.DataArtsUpdateHetuConnection(logger, subArgs)
		}
	case args.Fim != nil:
		address, mapped := cfg.Fim.Addresses[cfg.Fim.Address]
		if !mapped {
			address = cfg.Fim.Address
		}
		fimClient, err := fim.NewClient(address)
		if err != nil {
			logger.Fatal("unable to create FIM client.", logger.Args("error", err))
		}
		defer fimClient.Close()

		switch {
		case args.Fim.ResetUserPassword != nil:
			subArgs := args.Fim.ResetUserPassword
			subArgs.FimClient = fimClient
			subArgs.LoginUser = cfg.Fim.User
			subArgs.LoginPass = cfg.Fim.Password
			if subArgs.DefaultPass == "" {
				subArgs.DefaultPass = cfg.Fim.DefaultPassword
			}

			commands.FimResetUserPassword(logger, subArgs)
		}
	case args.Hive != nil:
		hiveServerClient, err := hive.NewHSClient(cfg.Hive.Url)
		if err != nil {
			logger.Fatal("unable to create HiveServer client.", logger.Args("error", err))
		}
		defer hiveServerClient.Close()

		if cfg.Hive.HostQualName != "" {
			os.Setenv("SERVICE_HOST_QUALIFIED", cfg.Hive.HostQualName)
		}

		switch {
		case args.Hive.BackupTable != nil:
			subArgs := args.Hive.BackupTable
			subArgs.HiveServerClient = hiveServerClient

			commands.HiveBackupTable(logger, subArgs)
		}
	case args.Iam != nil:
		logger.Debug("creating IAM client.")
		iamClient, err := iam.NewClient(cfg.AccessKey, cfg.SecretKey, cfg.SessionToken, cfg.Region)
		if err != nil {
			logger.Fatal("unable to create IAM client.", logger.Args("error", err))
		}

		switch {
		case args.Iam.ListUsers != nil:
			subArgs := args.Iam.ListUsers
			subArgs.DomainId = cfg.DomainId
			subArgs.IamClient = iamClient

			commands.IamListUsers(logger, subArgs)
		case args.Iam.ListGroups != nil:
			subArgs := args.Iam.ListGroups
			subArgs.DomainId = cfg.DomainId
			subArgs.IamClient = iamClient

			commands.IamListGroups(logger, subArgs)
		}
	case args.Mrs != nil:
		mrsClient, err := mrs.NewClient(cfg.AccessKey, cfg.SecretKey, cfg.SessionToken, cfg.Region, cfg.Mrs.ProxyAddress)
		if err != nil {
			logger.Fatal("unable to create MRS client.", logger.Args("error", err))
		}

		switch {
		case args.Mrs.DumpHetuClusters != nil:
			subArgs := args.Mrs.DumpHetuClusters
			subArgs.MrsClient = mrsClient
			subArgs.MrsClusterId = cfg.Mrs.ClusterId
			if subArgs.LoginUser == "" && cfg.Fim != nil {
				subArgs.LoginUser = cfg.Fim.User
			}
			address, mapped := cfg.Fim.Addresses[subArgs.FimAddress]
			if !mapped {
				address = subArgs.FimAddress
			}

			fimClient, err := fim.NewClient(address)
			if err != nil {
				logger.Fatal("unable to create FIM client.", logger.Args("error", err))
			}
			defer fimClient.Close()
			subArgs.FimClient = fimClient

			commands.MrsDumpHetuClusters(logger, subArgs)
		case args.Mrs.ListHetuTenants != nil:
			subArgs := args.Mrs.ListHetuTenants
			subArgs.MrsClient = mrsClient
			subArgs.MrsClusterId = cfg.Mrs.ClusterId
			if subArgs.LoginUser == "" && cfg.Fim != nil {
				subArgs.LoginUser = cfg.Fim.User
			}
			address, mapped := cfg.Fim.Addresses[subArgs.FimAddress]
			if !mapped {
				address = subArgs.FimAddress
			}

			fimClient, err := fim.NewClient(address)
			if err != nil {
				logger.Fatal("unable to create FIM client.", logger.Args("error", err))
			}
			defer fimClient.Close()
			subArgs.FimClient = fimClient

			commands.MrsListHetuTenants(logger, subArgs)
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

			commands.ObsAnalyze(logger, subArgs)
		case args.Obs.BatchRename != nil:
			subArgs := args.Obs.BatchRename
			subArgs.ObsClient = obsClient
			if !strings.HasSuffix(subArgs.Path, "/") {
				subArgs.Path += "/"
			}

			commands.ObsBatchRename(logger, subArgs)
		case args.Obs.BatchSetStorageClass != nil:
			subArgs := args.Obs.BatchSetStorageClass
			subArgs.ObsClient = obsClient

			commands.ObsBatchSetStorageClass(logger, subArgs)
		}
	case args.Ps != nil:
		switch {
		case args.Ps.AutoKill != nil:
			commands.PsAutoKill(logger, args.Ps.AutoKill)
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

			commands.UamDescribeUser(logger, subArgs)
		case args.Uam.ListMembers != nil:
			subArgs := args.Uam.ListMembers
			subArgs.BaseDn = cfg.Uam.BaseDN
			subArgs.UamClient = uamClient

			commands.UamListMembers(logger, subArgs)
		}
	case args.Yarn != nil:
		yarnClient, err := yarn.NewClient([]string(cfg.Yarn.RMAddress))
		if err != nil {
			logger.Fatal("unable to create YARN client.", logger.Args("error", err))
		}
		defer yarnClient.Close()

		switch {
		case args.Yarn.AutoKillApps != nil:
			subArgs := args.Yarn.AutoKillApps
			subArgs.YarnClient = yarnClient

			commands.YarnAutoKillApps(logger, subArgs)
		}
	}
}
