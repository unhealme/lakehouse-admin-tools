package cmd

const (
	DataArtsCreateHetuConnectionVersion = "2026.06.19-0"
	ObsAnalyzeVersion                   = "2026.06.22-0"
	ObsBatchRenameVersion               = "2026.06.22-0"
	ObsBatchSetStorageClassVersion      = "2026.06.22-0"
	UamDescribeUserVersion              = "2026.06.23-0"
	YarnAutoKillAppsVersion             = "2026.06.22-0"

	Version = "0.5.2"
)

var compVer = map[string]string{
	"dataarts-create-hetu-connection": DataArtsCreateHetuConnectionVersion,
	"obs-analyze":                     ObsAnalyzeVersion,
	"obs-batch-rename":                ObsBatchRenameVersion,
	"obs-batch-set-storage-class":     ObsBatchSetStorageClassVersion,
	"uam-describe-user":               UamDescribeUserVersion,
	"yarn-auto-kill-apps":             YarnAutoKillAppsVersion,
}
