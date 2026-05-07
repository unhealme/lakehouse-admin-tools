package obs

import (
	"fmt"
	rand "math/rand/v2"
	"path"
	"slices"
	"strings"
	"time"

	"github.com/pterm/pterm"
	"github.com/unhealme/lakehouse-admin-tools/cmd"
	"github.com/unhealme/lakehouse-admin-tools/internal"
	"github.com/unhealme/lakehouse-admin-tools/internal/obs"
)

type pathToRename struct{ before, after string }

func BatchRename(logger *internal.Logger, obsClient *obs.ObsClient, args *cmd.BatchRenameArgs) {
	logger.Debug("using batch rename args.", logger.Args(internal.ToArgs(*args)...))
	inputPath, err := obs.PathFromURI(args.Path)
	if err != nil {
		logger.Warn("skipping input due to error.", logger.Args("path", args.Path, "error", err))
		return
	}
	var paths []pathToRename
	for op := range obsClient.Walk(logger, inputPath.Bucket, inputPath.Key, 1, args.DirOnly) {
		if base, _ := strings.CutPrefix(op.Key, inputPath.Key); len(base) > 0 && !strings.HasPrefix(base, args.Prefix) {
			after := path.Join(inputPath.Key, args.Prefix+base)
			paths = append(paths, pathToRename{op.Key, after})
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
			func(path pathToRename) {
				if !args.DryRun {
					obsClient.RenameObject(logger, inputPath.Bucket, path.before, path.after)
				} else {
					logger.Info("renaming directory.", logger.Args("before", fmt.Sprintf("obs://%s/%s", inputPath.Bucket, path.before), "after", fmt.Sprintf("obs://%s/%s", inputPath.Bucket, path.after)))
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
