package config

import "github.com/unhealme/lakehouse-admin-tools/cmd"

const Version = "0.8.1"

var compVer = map[string]string{
	"dataarts-create-hetu-connection": cmd.DataArtsCreateHetuConnectionVersion,
	"mrs-list-hetu-tenants":           cmd.MrsListHetuTenantsVersion,
	"obs-analyze":                     cmd.ObsAnalyzeVersion,
	"obs-batch-rename":                cmd.ObsBatchRenameVersion,
	"obs-batch-set-storage-class":     cmd.ObsBatchSetStorageClassVersion,
	"ps-auto-kill":                    cmd.PsAutoKillVersion,
	"uam-describe-user":               cmd.UamDescribeUserVersion,
	"uam-list-members":                cmd.UamListMembersVersion,
	"yarn-auto-kill-apps":             cmd.YarnAutoKillAppsVersion,
}
