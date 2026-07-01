package cmd

import "github.com/unhealme/lakehouse-admin-tools/internal/obs"

type ObsBatchRenameArgs struct {
	Path        string `arg:"positional,required" placeholder:"PATH"`
	Prefix      string `arg:"positional,required" placeholder:"PREFIX"`
	Concurrency int    `arg:"-j,--" default:"2" help:"max job concurrency" placeholder:"NUM"`
	DirOnly     bool   `arg:"-d,--dir-only" help:"only rename directories"`
	DryRun      bool   `arg:"-,--dry-run" help:"simulate action without doing anything"`
	NoProg      bool   `arg:"-,--no-progress" help:"disable progress bar"`

	ObsClient *obs.ObsClient `arg:"-"`
}
