package cmd

import (
	"strings"

	"github.com/pterm/pterm"
	"github.com/unhealme/lakehouse-admin-tools/internal"
)

func DataArtsCreateHetuConnection(logger *internal.Logger, args *DataArtsCreateHetuConnectionArgs) {
	logger.Debug("using create hetu connection args.", logger.Args(internal.ToArgs(*args)...))
	for _, userPair := range args.UserPairs {
		userName, workspaceName, _ := strings.Cut(userPair, ":")
		logArgs := func(err error) []pterm.LoggerArgument {
			args := logger.Args("user", userName, "workspace", workspaceName)
			if err != nil {
				args = append(args, pterm.LoggerArgument{Key: "error", Value: err})
			}
			return args
		}

		logger.Debug("trying to get user information.", logArgs(nil))
		user, err := args.IamClient.GetUser(args.DomainId, userName, false)
		if err != nil {
			logger.Error("unable to get user information.", logArgs(err))
			continue
		}

		logger.Debug("trying to get workspace information.", logArgs(nil))
		workspace, err := args.DataArtsClient.GetWorkspaceFromName(args.InstanceId, workspaceName, false)
		if err != nil {
			logger.Error("unable to get workspace information.", logArgs(err))
			continue
		}

		logger.Debug("trying to create hetu connection.", logArgs(nil))
		if err := args.DataArtsClient.CreateHetuConnection(*workspace.Id, user.Name, args.AgentId, args.AgentName, args.HetuConfig); err != nil {
			logger.Error("unable to create hetu connection.", logArgs(err))
			continue
		}

		logger.Debug("trying to get hetu connection id.", logArgs(nil))
		connection, err := args.DataArtsClient.GetConnectionFromName(*workspace.Id, user.Name, true)
		if err != nil {
			logger.Error("unable to get hetu connection id.", logArgs(err))
			continue
		} else {
			logger.Info("successfully created hetu connection.", logArgs(nil))
		}

		logger.Debug("trying to assign connection permission.", logArgs(nil))
		if err := args.DataArtsClient.CreateResourcePermission(*workspace.Id, connection, user); err != nil {
			logger.Warn("unable to assign connection permission.", logArgs(err))
		} else {
			logger.Info("successfully assigned connection permission.", logArgs(nil))
		}
	}
}
