package main

import (
	"strings"

	arg "github.com/alexflint/go-arg"
	"github.com/unhealme/lakehouse-admin-tools/cmd"
	obs_cmd "github.com/unhealme/lakehouse-admin-tools/cmd/obs"
	"github.com/unhealme/lakehouse-admin-tools/internal"
	"github.com/unhealme/lakehouse-admin-tools/internal/obs"
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
	case args.OBS != nil:
		if args.OBS.Endpoint != nil {
			cfg.OBS.Endpoint = *args.OBS.Endpoint
		}

		obsClient, err := obs.NewClient(logger, cfg.OBS.Endpoint, cfg.AccessKey, cfg.SecretKey, cfg.SessionToken)
		if err != nil {
			logger.Fatal("unable to create OBS client.", logger.Args("error", err))
		}
		defer obsClient.Close()

		switch {
		case args.OBS.BatchRename != nil:
			if !strings.HasSuffix(args.OBS.BatchRename.Path, "/") {
				args.OBS.BatchRename.Path += "/"
			}
			obs_cmd.BatchRename(logger, obsClient, args.OBS.BatchRename)
		case args.OBS.BatchSetStorageClass != nil:
			obs_cmd.BatchSetStorageClass(logger, obsClient, args.OBS.BatchSetStorageClass)
		}
	}
}
