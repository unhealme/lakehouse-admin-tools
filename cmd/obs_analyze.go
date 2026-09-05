package cmd

import (
	"context"
	"strings"
	"time"

	"github.com/pterm/pterm"
	"github.com/unhealme/lakehouse-admin-tools/args"
	"github.com/unhealme/lakehouse-admin-tools/internal"
	"github.com/unhealme/lakehouse-admin-tools/internal/obs"
	"github.com/unhealme/lakehouse-admin-tools/utils"
	"go.uber.org/atomic"
)

const ObsAnalyzeVersion = "2026.09.05-1"

func ObsAnalyze(logger *pterm.Logger, args *args.ObsAnalyzeArgs) {
	logger.Debug("using analyze args.", logger.Args(internal.ToArgs(*args)...))
	if args.CsvOut != "" && args.CsvOut == args.JsonOut {
		logger.Fatal("unable to write csv and json output to the same file.")
	}

	if args.CsvOut != "" {
		if err := utils.OpenCsvWriter(args.CsvOut, obs.ObsPathAnalyzedHeader); err != nil {
			logger.Fatal("unable to open file to write.", logger.Args("file", args.CsvOut, "error", err))
		}
	}
	if args.JsonOut != "" {
		if err := utils.OpenJsonWriter(args.JsonOut); err != nil {
			logger.Fatal("unable to open file to write.", logger.Args("file", args.CsvOut, "error", err))
		}
	}
	defer utils.CloseOutput()

	type resultPath struct {
		raw    *obs.ObsPath
		input  obs.ObsPath
		result chan obs.ObsPathAnalyzed
	}

	var inputPaths []resultPath
	for _, uri := range args.Paths {
		inputPath, err := obs.PathFromURI(uri)
		if err != nil {
			logger.Warn("skipping path due to error.", logger.Args("path", uri, "error", err))
			continue
		}

		if args.Fixed || !strings.ContainsAny(inputPath.Key, obs.GlobToken) {
			inputPaths = append(
				inputPaths,
				resultPath{inputPath, *inputPath, make(chan obs.ObsPathAnalyzed, 1)},
			)
		} else {
			keys := args.ObsClient.Glob(logger, *inputPath)
			logger.Debug("glob path expanded.", logger.Args("totalKeys", len(keys)))
			if len(keys) < 1 {
				inputPaths = append(
					inputPaths,
					resultPath{inputPath, *inputPath, make(chan obs.ObsPathAnalyzed, 1)},
				)
			} else {
				for _, key := range keys {
					inputPaths = append(
						inputPaths,
						resultPath{inputPath, inputPath.WithKey(key), make(chan obs.ObsPathAnalyzed, 1)},
					)
				}
			}
		}
	}

	var (
		totalSize    atomic.Int64
		totalDirs    atomic.Int64
		totalFiles   atomic.Int64
		lastModified atomic.Int64

		slot = utils.NewSlot(args.Concurrency)
	)
	slot.DoNow(func() {
		var pathKey string
		pathExists := false
		for _, key := range inputPaths {
			if pathKey != "" && pathKey != key.raw.URI() {
				if !pathExists {
					pterm.Printf("%s: no such file or directory\n", pathKey)
				}
				pathExists = false
			}
			pathKey = key.raw.URI()

			stats := <-key.result
			if stats.Exists {
				pathExists = true
				utils.WriteOutput(stats)
				pterm.Printf(
					"obs://%s/%s: size: %d (%s), objects: %d (%d dirs, %d files), last modified: %s\n",
					stats.Bucket, stats.Key,
					stats.Size, utils.FormatSize(stats.Size),
					stats.DirCount+stats.FileCount,
					stats.DirCount, stats.FileCount,
					time.Unix(0, stats.LastModified*int64(time.Millisecond)).Format("2006-01-02 15:04:05.000"),
				)
				totalSize.Add(stats.Size)
				totalDirs.Add(stats.DirCount)
				totalFiles.Add(stats.FileCount)
				for {
					cur := lastModified.Load()
					if stats.LastModified <= cur {
						break
					}
					if lastModified.CompareAndSwap(cur, stats.LastModified) {
						break
					}
				}
				if strings.HasSuffix(stats.Key, "/") {
					totalDirs.Inc()
				}
			}
		}
		if !pathExists {
			pterm.Printf("%s: no such file or directory\n", pathKey)
		}
	})

	var prog *pterm.ProgressbarPrinter
	var progDone context.CancelFunc
	if !args.NoProg {
		var ctx context.Context
		ctx, progDone = context.WithCancel(context.Background())
		prog, _ = utils.NewProgressBar(ctx).WithTitle("Analyzing paths").
			WithTotal(len(inputPaths)).WithRemoveWhenDone(true).Start()
		_ = progDone
	}

	for i, path := range inputPaths {
		slot.Do(func() {
			inputPaths[i].result <- args.ObsClient.Analyze(logger, slot, args.MinChunks, path.input)
			if prog != nil {
				prog.Increment()
			}
		})
	}
	slot.Wait()
	if prog != nil {
		progDone()
	}

	if args.Summarize {
		pterm.Println()
		pterm.Printf(
			"Total size: %d (%s), objects: %d (%d dirs, %d files), last modified: %s\n",
			totalSize.Load(), utils.FormatSize(totalSize.Load()),
			totalDirs.Load()+totalFiles.Load(),
			totalDirs.Load(), totalFiles.Load(),
			time.Unix(0, lastModified.Load()*int64(time.Millisecond)).Format("2006-01-02 15:04:05.000"),
		)
	}
}
