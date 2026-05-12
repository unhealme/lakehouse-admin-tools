package cmd

const (
	ObsBatchRenameVersion          = "2026.05.07-0"
	ObsBatchSetStorageClassVersion = "2026.05.07-0"
	YarnAutoKillAppsVersion        = "2026.05.13-0"

	Version = "0.2.0"
)

var compVer = map[string]string{
	"obs-batch-rename":            ObsBatchRenameVersion,
	"obs-batch-set-storage-class": ObsBatchSetStorageClassVersion,
	"yarn-auto-kill-apps":         YarnAutoKillAppsVersion,
}
