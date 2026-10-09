package config

import "github.com/unhealme/lakehouse-admin-tools/pkg/commands"

const Version = "0.14.9"

var compVer = map[string]string{
	"dataarts-create-hetu-connection": commands.DataArtsCreateHetuConnectionVersion,
	"dataarts-update-hetu-connection": commands.DataArtsUpdateHetuConnectionVersion,
	"fim-reset-user-password":         commands.FimResetUserPasswordVersion,
	"hive-backup-table":               commands.HiveBackupTableVersion,
	"iam-list-groups":                 commands.IamListGroupsVersion,
	"iam-list-users":                  commands.IamListUsersVersion,
	"mrs-dump-hetu-clusters":          commands.MrsDumpHetuClustersVersion,
	"mrs-list-hetu-tenants":           commands.MrsListHetuTenantsVersion,
	"obs-analyze":                     commands.ObsAnalyzeVersion,
	"obs-batch-rename":                commands.ObsBatchRenameVersion,
	"obs-batch-set-storage-class":     commands.ObsBatchSetStorageClassVersion,
	"ps-auto-kill":                    commands.PsAutoKillVersion,
	"uam-describe-user":               commands.UamDescribeUserVersion,
	"uam-list-members":                commands.UamListMembersVersion,
	"yarn-auto-kill-apps":             commands.YarnAutoKillAppsVersion,
	"yarn-list-apps":                  commands.YarnListAppsVersion,
}
