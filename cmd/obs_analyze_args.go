package cmd

import "github.com/unhealme/lakehouse-admin-tools/internal/obs"

type ObsAnalyzeArgs struct {
	Paths   []string `arg:"positional,required" placeholder:"PATH"`
	DirOnly bool     `arg:"-d,--dir-only" help:"only analyze directories"`

	ObsClient *obs.ObsClient `arg:"-"`
}
