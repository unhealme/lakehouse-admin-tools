package main

import (
	"strings"

	arg "github.com/alexflint/go-arg"
	"github.com/unhealme/lakehouse-admin-tools/cmd"
	das_cmd "github.com/unhealme/lakehouse-admin-tools/cmd/dataarts-studio"
	obs_cmd "github.com/unhealme/lakehouse-admin-tools/cmd/obs"
	uam_cmd "github.com/unhealme/lakehouse-admin-tools/cmd/uam"
	yarn_cmd "github.com/unhealme/lakehouse-admin-tools/cmd/yarn"
	"github.com/unhealme/lakehouse-admin-tools/internal"
	das "github.com/unhealme/lakehouse-admin-tools/internal/dataarts-studio"
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
		dasClient, err := das.NewClient(cfg.AccessKey, cfg.SecretKey, cfg.SessionToken, cfg.Region)
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

			das_cmd.CreateHetuConnection(logger, args.DataArts.CreateHetuConnection)
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

			obs_cmd.Analyze(logger, args.Obs.Analyze)
		case args.Obs.BatchRename != nil:
			args.Obs.BatchRename.ObsClient = obsClient
			if !strings.HasSuffix(args.Obs.BatchRename.Path, "/") {
				args.Obs.BatchRename.Path += "/"
			}

			obs_cmd.BatchRename(logger, args.Obs.BatchRename)
		case args.Obs.BatchSetStorageClass != nil:
			args.Obs.BatchSetStorageClass.ObsClient = obsClient

			obs_cmd.BatchSetStorageClass(logger, args.Obs.BatchSetStorageClass)
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

			uam_cmd.DescribeUser(logger, args.Uam.DescribeUser)
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

			yarn_cmd.AutoKillApps(logger, args.Yarn.AutoKillApps)
		}
	}
}
