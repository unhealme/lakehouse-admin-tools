package cmd

const (
	DataArtsCreateHetuConnectionVersion = "2026.06.19-0"
	ObsAnalyzeVersion                   = "2026.06.22-0"
	ObsBatchRenameVersion               = "2026.06.22-0"
	ObsBatchSetStorageClassVersion      = "2026.06.22-0"
	PsAutoKillVersion                   = "2026.06.30-1"
	UamDescribeUserVersion              = "2026.07.01-0"
	UamListMembersVersion               = "2026.06.24-1"
	YarnAutoKillAppsVersion             = "2026.06.22-0"

	Version = "0.7.1"
)

var compVer = map[string]string{
	"dataarts-create-hetu-connection": DataArtsCreateHetuConnectionVersion,
	"obs-analyze":                     ObsAnalyzeVersion,
	"obs-batch-rename":                ObsBatchRenameVersion,
	"obs-batch-set-storage-class":     ObsBatchSetStorageClassVersion,
	"ps-auto-kill":                    PsAutoKillVersion,
	"uam-describe-user":               UamDescribeUserVersion,
	"uam-list-members":                UamListMembersVersion,
	"yarn-auto-kill-apps":             YarnAutoKillAppsVersion,
}
