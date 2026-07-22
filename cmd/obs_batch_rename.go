package cmd

import (
	rand "math/rand/v2"
	"path"
	"slices"
	"strings"
	"time"

	"github.com/pterm/pterm"
	"github.com/unhealme/lakehouse-admin-tools/internal"
	"github.com/unhealme/lakehouse-admin-tools/internal/obs"
)

const ObsBatchRenameVersion = "2026.07.11-0"

func ObsBatchRename(logger *pterm.Logger, args *ObsBatchRenameArgs) {
	logger.Debug("using batch rename args.", logger.Args(internal.ToArgs(*args)...))
	inputPath, err := obs.PathFromURI(args.Path)
	if err != nil {
		logger.Warn("skipping input due to error.", logger.Args("path", args.Path, "error", err))
		return
	}

	type pathToRename struct {
		before   obs.ObsPath
		keyAfter string
	}
	var paths []pathToRename
	var total int
	for op := range args.ObsClient.Walk(logger, *inputPath, 1, args.DirOnly) {
		if base := strings.TrimPrefix(op.Key, inputPath.Key); len(base) > 0 && !strings.HasPrefix(base, args.Prefix) {
			before := obs.NewObsPath(op.Bucket, op.Key)
			paths = append(paths, pathToRename{before, path.Join(inputPath.Key, args.Prefix+base)})
			total += 1
		}
	}

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
					args.ObsClient.RenameObject(logger, path.before, path.keyAfter)
				} else {
					logger.Info("renaming directory.", logger.Args("before", path.before.URI(), "after", path.before.WithKey(path.keyAfter).URI()))
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
