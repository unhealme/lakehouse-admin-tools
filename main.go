package main

import (
	"strings"

	arg "github.com/alexflint/go-arg"
	"github.com/unhealme/lakehouse-admin-tools/cmd"
	obs_cmd "github.com/unhealme/lakehouse-admin-tools/cmd/obs"
	yarn_cmd "github.com/unhealme/lakehouse-admin-tools/cmd/yarn"
	"github.com/unhealme/lakehouse-admin-tools/internal"
	"github.com/unhealme/lakehouse-admin-tools/internal/obs"
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

	switch {
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
		case args.Obs.BatchRename != nil:
			if !strings.HasSuffix(args.Obs.BatchRename.Path, "/") {
				args.Obs.BatchRename.Path += "/"
			}
			obs_cmd.BatchRename(logger, obsClient, args.Obs.BatchRename)
		case args.Obs.BatchSetStorageClass != nil:
			obs_cmd.BatchSetStorageClass(logger, obsClient, args.Obs.BatchSetStorageClass)
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
			yarn_cmd.AutoKillApps(logger, yarnClient, args.Yarn.AutoKillApps)
		}
	}
}
