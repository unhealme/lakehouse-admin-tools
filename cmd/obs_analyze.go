package cmd

import (
	"strings"
	"sync"

	"github.com/pterm/pterm"
	"github.com/unhealme/lakehouse-admin-tools/internal"
	"github.com/unhealme/lakehouse-admin-tools/internal/obs"
	utils "github.com/unhealme/lakehouse-admin-tools/internal/obs/analyze-utils"
	"go.uber.org/atomic"
)

const ObsAnalyzeVersion = "2026.07.22-0"

func ObsAnalyze(logger *pterm.Logger, args *ObsAnalyzeArgs) {
	logger.Debug("using analyze args.", logger.Args(internal.ToArgs(*args)...))
	if args.CsvOut != "" && args.CsvOut == args.JsonOut {
		logger.Fatal("unable to write csv and json output to the same file.")
	}

	var outputWriter utils.OutputWriter
	if args.CsvOut != "" {
		if err := outputWriter.OpenCsvWriter(args.CsvOut); err != nil {
			logger.Fatal("unable to open file to write.", logger.Args("file", args.CsvOut, "error", err))
		}
	}
	if args.JsonOut != "" {
		if err := outputWriter.OpenJsonWriter(args.JsonOut); err != nil {
			logger.Fatal("unable to open file to write.", logger.Args("file", args.CsvOut, "error", err))
		}
	}
	defer outputWriter.Close()

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
			inputPaths = append(inputPaths,
				resultPath{inputPath, *inputPath, make(chan obs.ObsPathAnalyzed)},
			)
		} else {
			for _, key := range args.ObsClient.Glob(logger, *inputPath) {
				inputPaths = append(inputPaths,
					resultPath{inputPath, inputPath.WithKey(key), make(chan obs.ObsPathAnalyzed)},
				)
			}
		}
	}

	var (
		totalSize  atomic.Int64
		totalDirs  atomic.Int64
		totalFiles atomic.Int64

		wg sync.WaitGroup
	)
	wg.Go(func() {
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
				outputWriter.Write(stats)
				totalSize.Add(stats.Size)
				totalDirs.Add(int64(stats.DirCount))
				totalFiles.Add(int64(stats.FileCount))
				if strings.HasSuffix(stats.Key, "/") {
					totalDirs.Inc()
				} else {
					totalFiles.Inc()
				}
			}
		}
		if !pathExists {
			pterm.Printf("%s: no such file or directory\n", pathKey)
		}
	})

	var prog *pterm.ProgressbarPrinter
	if !args.NoProg {
		prog, _ = internal.NewProgressBar().WithTitle("Analyzing paths").WithTotal(len(inputPaths)).WithRemoveWhenDone(true).Start()
	}

	sem := make(chan internal.EmptyType, max(args.Concurrency, 1))
	for i, path := range inputPaths {
		sem <- internal.Empty
		wg.Go(func() {
			inputPaths[i].result <- args.ObsClient.Analyze(logger, path.input)
			if prog != nil {
				prog.Increment()
			}
			<-sem
		})
	}
	wg.Wait()
	if prog != nil {
		prog.Stop()
	}

	if args.Summarize {
		pterm.Println()
		pterm.Printf(
			"Total size: %d (%s), objects: %d (%d dirs, %d files)\n",
			totalSize.Load(), internal.FormatSize(totalSize.Load()),
			totalDirs.Load()+totalFiles.Load(),
			totalDirs.Load(), totalFiles.Load(),
		)
	}
}
