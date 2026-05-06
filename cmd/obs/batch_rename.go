package obs

import (
	rand "math/rand/v2"
	"slices"
	"strings"
	"time"

	"github.com/pterm/pterm"
	"github.com/unhealme/lakehouse-admin-tools/cmd"
	"github.com/unhealme/lakehouse-admin-tools/internal"
	"github.com/unhealme/lakehouse-admin-tools/internal/obs"
)

func BatchRename(logger *internal.Logger, obsClient *obs.ObsClient, args *cmd.BatchRenameArgs) {
	logger.Debug("using batch rename args.", logger.Args(internal.ToArgs(*args)...))
	var paths [][2]string
	for op := range obsClient.Walk(args.Bucket, args.Path, 1, args.DirOnly) {
		if name, _ := strings.CutPrefix(op.Key, args.Path); len(name) > 0 && !strings.HasPrefix(name, args.Prefix) {
			after := args.Path + args.Prefix + name
			paths = append(paths, [2]string{op.Key, after})
		}
	}

	total := len(paths)
	logger.Info("all obs paths fetched.", logger.Args("path to rename", total))
	if total > 0 {
		var prog *pterm.ProgressbarPrinter
		if !args.NoProg {
			prog, _ = internal.NewProgressBar().WithTitle("Renaming paths").WithTotal(total).Start()
			defer prog.Stop()
		}
		internal.ParallelMap(
			func(path [2]string) {
				if !args.DryRun {
					obsClient.RenameObject(args.Bucket, path[0], path[1])
				} else {
					logger.Info("renaming directory.", logger.Args("before", path[0], "after", path[1]))
					time.Sleep(200 + rand.N(300*time.Millisecond))
				}
				if prog != nil {
					prog.Increment()
				}
			},
			slices.Values(paths),
			args.Concurrency,
		)
	}
	logger.Info("rename obs paths done.")
}
