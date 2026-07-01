package cmd

import "github.com/unhealme/lakehouse-admin-tools/internal/obs"

type ObsBatchSetStorageClassArgs struct {
	InputFiles  []string `arg:"positional,required" placeholder:"FILE"`
	Concurrency int      `arg:"-j,--" default:"2" help:"max job concurrency" placeholder:"NUM"`
	DryRun      bool     `arg:"-,--dry-run" help:"simulate action without doing anything"`
	NoProg      bool     `arg:"-,--no-progress" help:"disable progress bar"`

	ObsClient *obs.ObsClient `arg:"-"`
}
