package cmd

import "github.com/unhealme/lakehouse-admin-tools/internal"

type PsAutoKillArgs struct {
	RuntimeThreshold internal.Duration `arg:"-t,--runtime-threshold,required" help:"kill process if runtime is greater than DUR" placeholder:"DUR"`
	MemoryThreshold  string            `arg:"-m,--memory-threshold,required" help:"kill process if memory usage is greter than NUM bytes" placeholder:"NUM"`
	ExcludeUsers     string            `arg:"-u,--exclude-users" help:"exclude user process from auto kill" placeholder:"USER[,USER...]"`
	DryRun           bool              `arg:"-,--dry-run" help:"simulate action without doing anything"`
}
