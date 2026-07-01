package main

import (
	"strings"

	arg "github.com/alexflint/go-arg"
	"github.com/unhealme/lakehouse-admin-tools/cmd"
	"github.com/unhealme/lakehouse-admin-tools/internal"
	"github.com/unhealme/lakehouse-admin-tools/internal/dataarts"
	"github.com/unhealme/lakehouse-admin-tools/internal/iam"
	"github.com/unhealme/lakehouse-admin-tools/internal/obs"
	"github.com/unhealme/lakehouse-admin-tools/internal/uam"
	"github.com/unhealme/lakehouse-admin-tools/internal/yarn"
)

var logger = internal.DefaultLogger()

func main() {
	var args cmd.Arguments
	arg.MustParse(&args)
	if args.Verbose {
		logger = logger.WithLevel(internal.LogLevelDebug)
	}
	logger.Debug("arguments parsed.", logger.Args(internal.ToArgs(args)...))

	cfg := internal.GetConfig(logger, args.ConfigFile)
	if args.AccessKey != "" {
		cfg.AccessKey = args.AccessKey
	}
	if args.SecretKey != "" {
		cfg.SecretKey = args.SecretKey
	}
	if args.SessionToken != "" {
		cfg.SessionToken = args.SessionToken
	}
	if args.DomainId != "" {
		cfg.DomainId = args.DomainId
	}
	if args.Region != "" {
		cfg.Region = args.Region
	}

	switch {
	case args.DataArts != nil:
		if args.DataArts.InstanceId != "" {
			cfg.DataArts.InstanceId = args.DataArts.InstanceId
		}

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
			args.DataArts.CreateHetuConnection.DomainId = cfg.DomainId
			args.DataArts.CreateHetuConnection.InstanceId = cfg.DataArts.InstanceId
			args.DataArts.CreateHetuConnection.DataArtsClient = dasClient
			args.DataArts.CreateHetuConnection.IamClient = iamClient
			args.DataArts.CreateHetuConnection.HetuConfig = &cfg.DataArts.HetuConfig
			if args.DataArts.CreateHetuConnection.AgentId == "" {
				args.DataArts.CreateHetuConnection.AgentId = cfg.DataArts.Agent.Id
			}
			if args.DataArts.CreateHetuConnection.AgentName == "" {
				args.DataArts.CreateHetuConnection.AgentName = cfg.DataArts.Agent.Name
			}

			cmd.DataArtsCreateHetuConnection(logger, args.DataArts.CreateHetuConnection)
		}
	case args.Obs != nil:
		if args.Obs.Endpoint != "" {
			cfg.Obs.Endpoint = args.Obs.Endpoint
		}

		obsClient, err := obs.NewClient(cfg.Obs.Endpoint, cfg.AccessKey, cfg.SecretKey, cfg.SessionToken)
		if err != nil {
			logger.Fatal("unable to create OBS client.", logger.Args("error", err))
		}
		defer obsClient.Close()

		switch {
		case args.Obs.Analyze != nil:
			args.Obs.Analyze.ObsClient = obsClient

			cmd.ObsAnalyze(logger, args.Obs.Analyze)
		case args.Obs.BatchRename != nil:
			args.Obs.BatchRename.ObsClient = obsClient
			if !strings.HasSuffix(args.Obs.BatchRename.Path, "/") {
				args.Obs.BatchRename.Path += "/"
			}

			cmd.ObsBatchRename(logger, args.Obs.BatchRename)
		case args.Obs.BatchSetStorageClass != nil:
			args.Obs.BatchSetStorageClass.ObsClient = obsClient

			cmd.ObsBatchSetStorageClass(logger, args.Obs.BatchSetStorageClass)
		}
	case args.Ps != nil:
		switch {
		case args.Ps.AutoKill != nil:
			cmd.PsAutoKill(logger, args.Ps.AutoKill)
		}
	case args.Uam != nil:
		if args.Uam.Url != "" {
			cfg.Uam.Url = args.Uam.Url
		}
		if args.Uam.User != "" {
			cfg.Uam.User = args.Uam.User
		}
		if args.Uam.Password != "" {
			cfg.Uam.Password = args.Uam.Password
		}
		if args.Uam.BaseDN != "" {
			cfg.Uam.BaseDN = args.Uam.BaseDN
		}
		if args.Uam.GroupBase != "" {
			cfg.Uam.GroupBase = args.Uam.GroupBase
		}
		if args.Uam.Domain != "" {
			cfg.Uam.Domain = args.Uam.Domain
		}

		uamClient, err := uam.NewClient(
			logger, cfg.Uam.Url, cfg.Uam.User, cfg.Uam.Password,
			cfg.Uam.BaseDN, cfg.Uam.GroupBase, cfg.Uam.Domain, cfg.Uam.Realm,
		)
		if err != nil {
			logger.Fatal("unable to create UAM client.", logger.Args("error", err))
		}
		defer uamClient.Close()

		switch {
		case args.Uam.DescribeUser != nil:
			args.Uam.DescribeUser.UamClient = uamClient

			cmd.UamDescribeUser(logger, args.Uam.DescribeUser)
		case args.Uam.ListMembers != nil:
			args.Uam.ListMembers.UamClient = uamClient

			cmd.UamListMembers(logger, args.Uam.ListMembers)
		}
	case args.Yarn != nil:
		if args.Yarn.RMAddress != "" {
			cfg.Yarn.RMAddress = args.Yarn.RMAddress
		}

		yarnClient, err := yarn.NewClient(cfg.Yarn.RMAddress)
		if err != nil {
			logger.Fatal("unable to create YARN client.", logger.Args("error", err))
		}
		defer yarnClient.CloseIdleConnections()

		switch {
		case args.Yarn.AutoKillApps != nil:
			args.Yarn.AutoKillApps.YarnClient = yarnClient

			cmd.YarnAutoKillApps(logger, args.Yarn.AutoKillApps)
		}
	}
}
