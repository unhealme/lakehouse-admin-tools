package cmd

import (
	"encoding/csv"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/pterm/pterm"
	"github.com/unhealme/lakehouse-admin-tools/internal"
	"github.com/unhealme/lakehouse-admin-tools/internal/hetu"
)

const MrsListHetuTenantsVersion = "2026.07.07-0"

var headerRow = []string{
	"Time",
	"Tenant",
	"Ids",
	"Vcores",
	"Memory",
	"Running",
	"Stopped",
	"Error",
}

func MrsListHetuTenants(logger *pterm.Logger, args *MrsListHetuTenantsArgs) {
	logger.Debug("using list tenants args.", logger.Args(internal.ToArgs(*args)...))

	resp, err := args.MrsClient.GetClusterManagerToken(args.MrsClusterId)
	if err != nil {
		logger.Fatal("unable to get MRS token.", logger.Args("error", err))
	}
	if err := args.FimClient.Login(args.LoginUser, resp.Token); err != nil {
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
	writer := csv.NewWriter(outFile)
	defer writer.Flush()
	if !args.NoHeader {
		writer.Write(headerRow)
	}
	now := time.Now().In(time.FixedZone("UTC+7", 7*3600))
	now = now.Add(-time.Duration(now.Minute()) * time.Minute).Add(-time.Duration(now.Second()) * time.Second)
	for tenant := range hetuClient.IterTenantInfo(logger) {
		writer.Write([]string{
			strconv.FormatInt(now.Unix(), 10),
			tenant.Tenant,
			strings.Join(tenant.ClusterIds, ", "),
			strconv.FormatInt(int64(tenant.TotalVcores), 10),
			internal.FormatSize(int64(tenant.TotalMemory * 1024 * 1024)),
			strconv.FormatInt(int64(tenant.RunningCount), 10),
			strconv.FormatInt(int64(tenant.StoppedCount), 10),
			strconv.FormatInt(int64(tenant.ErrorCount), 10),
		})
	}
}
