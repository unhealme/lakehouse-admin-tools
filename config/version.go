package config

import "github.com/unhealme/lakehouse-admin-tools/cmd"

const Version = "0.12.1"

var compVer = map[string]string{
	"dataarts-create-hetu-connection": cmd.DataArtsCreateHetuConnectionVersion,
	"dataarts-update-hetu-connection": cmd.DataArtsUpdateHetuConnectionVersion,
	"fim-reset-user-password":         cmd.FimResetUserPasswordVersion,
	"iam-list-groups":                 cmd.IamListGroupsVersion,
	"iam-list-users":                  cmd.IamListUsersVersion,
	"mrs-dump-hetu-clusters":          cmd.MrsDumpHetuClustersVersion,
	"mrs-list-hetu-tenants":           cmd.MrsListHetuTenantsVersion,
	"obs-analyze":                     cmd.ObsAnalyzeVersion,
	"obs-batch-rename":                cmd.ObsBatchRenameVersion,
	"obs-batch-set-storage-class":     cmd.ObsBatchSetStorageClassVersion,
	"ps-auto-kill":                    cmd.PsAutoKillVersion,
	"uam-describe-user":               cmd.UamDescribeUserVersion,
	"uam-list-members":                cmd.UamListMembersVersion,
	"yarn-auto-kill-apps":             cmd.YarnAutoKillAppsVersion,
}
