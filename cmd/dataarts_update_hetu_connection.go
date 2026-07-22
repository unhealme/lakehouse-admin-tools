package cmd

import (
	"strings"

	"github.com/goccy/go-json"
	"github.com/pterm/pterm"
	"github.com/unhealme/lakehouse-admin-tools/internal"
)

const DataArtsUpdateHetuConnectionVersion = "2026.07.18-0"

func DataArtsUpdateHetuConnection(logger *pterm.Logger, args *DataArtsUpdateHetuConnectionArgs) {
	logger.Debug("using update hetu connection args.", logger.Args(internal.ToArgs(*args)...))
	tenantDev, tenantProd, hasDev := strings.Cut(args.Tenant, ":")
	if !hasDev {
		tenantProd, tenantDev = tenantDev, tenantProd
	}

	for _, userPair := range args.UserPairs {
		userName, workspaceName, _ := strings.Cut(userPair, ":")
		logArgs := func(err error) []pterm.LoggerArgument {
			args := logger.Args("user", userName, "workspace", workspaceName)
			if err != nil {
				args = append(args, pterm.LoggerArgument{Key: "error", Value: err})
			}
			return args
		}

		logger.Debug("trying to get workspace information.", logArgs(nil))
		workspace, err := args.DataArtsClient.GetWorkspaceFromName(args.InstanceId, workspaceName, false)
		if err != nil {
			logger.Error("unable to get workspace information.", logArgs(err))
			continue
		}

		logger.Debug("trying to get hetu connection id.", logArgs(nil))
		connectionBasic, err := args.DataArtsClient.GetConnectionFromName(*workspace.Id, userName, true)
		if err != nil {
			logger.Error("unable to get hetu connection id.", logArgs(err))
			continue
		}

		logger.Debug("trying to get hetu connection config.", logArgs(nil))
		connection, err := args.DataArtsClient.GetConnectionConfig(*workspace.Id, *connectionBasic.DwId)
		if err != nil {
			logger.Error("unable to get hetu connection config.", logArgs(err))
			continue
		}
		config := (*connection.DwConfig).(map[string]any)

		configEncoded, _ := json.Marshal(config)
		logger.Debug("got hetu config.", logArgs(nil), logger.Args("config", string(configEncoded)))

		configProperties := make(map[string]string)
		if properties, ok := config["properties"]; ok {
			json.Unmarshal([]byte(properties.(string)), &configProperties)
		}

		hetuConfDev := args.HetuConfig
		emptyBool := false
		hetuConfDev.RememberPassword = &emptyBool
		hetuConfProd := hetuConfDev

		if tenantDev != "" {
			prop := configProperties
			prop["tenant"] = tenantDev
			propEncoded, _ := json.Marshal(prop)
			propEncodedStr := string(propEncoded)
			hetuConfDev.Properties = &propEncodedStr
		}

		if tenantProd != "" {
			prop := configProperties
			prop["tenant"] = tenantProd
			propEncoded, _ := json.Marshal(prop)
			propEncodedStr := string(propEncoded)
			hetuConfProd.Properties = &propEncodedStr
		}

		logger.Debug("trying to update hetu connection.", logArgs(nil))
		hetuConfDevInter := any(hetuConfDev)
		hetuConfProdInter := any(hetuConfProd)
		hetuConfDevEncoded, _ := json.Marshal(hetuConfDev)
		hetuConfProdEncoded, _ := json.Marshal(hetuConfProd)
		logger.Debug("using hetu config.", logArgs(nil), logger.Args("devConfig", string(hetuConfDevEncoded), "prodConfig", string(hetuConfProdEncoded)))
		if err := args.DataArtsClient.UpdateHetuConnection(*workspace.Id, connection, &hetuConfDevInter, &hetuConfProdInter); err != nil {
			logger.Warn("unable to update hetu connection.", logArgs(err))
		} else {
			logger.Info("successfully updated hetu connection.", logArgs(nil))
		}
	}
}
