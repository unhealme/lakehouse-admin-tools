package commands

import (
	"encoding/csv"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/pterm/pterm"
	"github.com/unhealme/lakehouse-admin-tools/internal"
	"github.com/unhealme/lakehouse-admin-tools/internal/clients/hetu"
	"github.com/unhealme/lakehouse-admin-tools/pkg/arguments"
	"github.com/unhealme/lakehouse-admin-tools/pkg/utils"
)

const MrsListHetuTenantsVersion = "2026.08.05-0"

func MrsListHetuTenants(logger *pterm.Logger, args *arguments.MrsListHetuTenantsArgs) {
	logger.Debug("using list hetu tenants args.", logger.Args(internal.ToArgs(*args)...))

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
		if outFile, err = os.Create(args.OutputFile); err != nil {
			logger.Fatal("unable to open file to write.", logger.Args("file", args.OutputFile, "error", err))
		}
		defer outFile.Close()
	}
	writer := csv.NewWriter(outFile)
	defer writer.Flush()
	if !args.NoHeader {
		if err := writer.Write([]string{
			"Time",
			"Tenant",
			"Ids",
			"Vcores",
			"Memory",
			"Running",
			"Stopped",
			"Error",
		}); err != nil {
			panic(err)
		}
	}
	now := time.Now().In(time.FixedZone("UTC+7", 7*3600))
	now = now.Add(-time.Duration(now.Minute()) * time.Minute).Add(-time.Duration(now.Second()) * time.Second)
	for tenant := range hetuClient.IterTenantInfo(logger) {
		if err := writer.Write([]string{
			strconv.FormatInt(now.Unix(), 10),
			tenant.Tenant,
			strings.Join(tenant.ClusterIds, ", "),
			strconv.FormatInt(int64(tenant.TotalVcores), 10),
			utils.FormatSize(int64(tenant.TotalMemory * 1024 * 1024)),
			strconv.FormatInt(int64(tenant.RunningCount), 10),
			strconv.FormatInt(int64(tenant.StoppedCount), 10),
			strconv.FormatInt(int64(tenant.ErrorCount), 10),
		}); err != nil {
			panic(err)
		}
	}
}
