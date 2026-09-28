package commands

import (
	"encoding/json"
	"os"

	"github.com/pterm/pterm"
	"github.com/unhealme/lakehouse-admin-tools/internal"
	"github.com/unhealme/lakehouse-admin-tools/internal/clients/hetu"
	"github.com/unhealme/lakehouse-admin-tools/pkg/arguments"
)

const MrsDumpHetuClustersVersion = "2026.08.06-0"

func MrsDumpHetuClusters(logger *pterm.Logger, args *arguments.MrsDumpHetuClustersArgs) {
	logger.Debug("using dump hetu clusters args.", logger.Args(internal.ToArgs(*args)...))

	resp, err := args.MrsClient.GetClusterManagerToken(args.MrsClusterId)
	if err != nil {
		logger.Fatal("unable to get MRS token.", logger.Args("error", err))
	}
	if err := args.FimClient.MrsLogin(args.LoginUser, resp.Token); err != nil {
		logger.Fatal("unable to login to FIM.", logger.Args("error", err))
	}

	hetuAuth, err := args.FimClient.GetHetuEngineAuth(logger, args.FimClusterId)
	if err != nil {
		logger.Fatal("unable to get Hetu auth.", logger.Args("error", err))
	}

	hetuClient := hetu.NewClient(hetuAuth)
	if err := hetuClient.GetToken(); err != nil {
		logger.Fatal("unable to get Hetu token.", logger.Args("error", err))
	}
	defer hetuClient.Close()

	outFile := os.Stdout
	if args.OutputFile != "" {
		var err error
		if outFile, err = os.OpenFile(
			args.OutputFile,
			os.O_WRONLY|os.O_CREATE|os.O_TRUNC,
			0o644,
		); err != nil {
			logger.Fatal("unable to open file to write.", logger.Args("file", args.OutputFile, "error", err))
		}
		defer outFile.Close()
	}
	clusters, err := hetuClient.ClustersRaw()
	if err != nil {
		logger.Fatal("unable to get Hetu clusters.", logger.Args("error", err))
	}
	result, _ := json.Marshal(clusters)
	if _, err := outFile.Write(result); err != nil {
		panic(err)
	}
}
