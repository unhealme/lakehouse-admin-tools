package config

import "github.com/unhealme/lakehouse-admin-tools/cmd"

const Version = "0.7.4"

var compVer = map[string]string{
	"dataarts-create-hetu-connection": cmd.DataArtsCreateHetuConnectionVersion,
	"obs-analyze":                     cmd.ObsAnalyzeVersion,
	"obs-batch-rename":                cmd.ObsBatchRenameVersion,
	"obs-batch-set-storage-class":     cmd.ObsBatchSetStorageClassVersion,
	"ps-auto-kill":                    cmd.PsAutoKillVersion,
	"uam-describe-user":               cmd.UamDescribeUserVersion,
	"uam-list-members":                cmd.UamListMembersVersion,
	"yarn-auto-kill-apps":             cmd.YarnAutoKillAppsVersion,
}
