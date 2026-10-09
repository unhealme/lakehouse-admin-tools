package commands

import (
	"os"
	"strings"
	"time"

	"github.com/pterm/pterm"
	"github.com/unhealme/lakehouse-admin-tools/internal"
	"github.com/unhealme/lakehouse-admin-tools/internal/clients/obs"
	"github.com/unhealme/lakehouse-admin-tools/internal/logger"
	"github.com/unhealme/lakehouse-admin-tools/pkg/arguments"
	"github.com/unhealme/lakehouse-admin-tools/pkg/utils"
)

const ObsAnalyzeVersion = "2026.10.09-0"

func ObsAnalyze(args *arguments.ObsAnalyzeArgs) {
	logger.Debug("using analyze args.", logger.Args(internal.StructToArgs(args)...))
	if args.CsvOut != "" && args.CsvOut == args.JsonOut {
		logger.Fatal("unable to write csv and json output to the same file.")
	}

	pathInputs := args.Paths
	if args.InputFile != "" || len(args.Paths) < 1 {
		r := os.Stdin
		if args.InputFile != "" && args.InputFile != "-" {
			var err error
			r, err = os.Open(args.InputFile)
			if err != nil {
				logger.Fatal("unable to read input file.", logger.Args("file", args.InputFile))
			}
		}
		re := make(chan error, 1)
		for i := range utils.IterLinesSeq(r, args.InputSep, re) {
			if t := strings.TrimSpace(i); t != "" {
				pathInputs = append(pathInputs, strings.TrimSpace(i))
			}
		}
		if err := <-re; err != nil {
			logger.Fatal("unable to read input file.", logger.Args("file", args.InputFile, "error", err))
		}
		if r != os.Stdin {
			r.Close()
		}
	}
	if len(pathInputs) < 1 {
		return
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

	var inputs []obs.ObsPath
	for _, uri := range pathInputs {
		path, err := obs.PathFromURI(uri)
		if err != nil {
			logger.Warn("skipping path due to error.", logger.Args("path", uri, "error", err))
			continue
		}

		if args.Fixed || !strings.ContainsAny(path.Key, obs.GlobToken) {
			inputs = append(inputs, *path)
		} else {
			keys := args.ObsClient.Glob(*path)
			logger.Debug("glob path expanded.", logger.Args("totalKeys", len(keys)))
			if len(keys) < 1 {
				logger.Error(path.URI() + ": no such file or directory")
			} else {
				for _, key := range keys {
					inputs = append(inputs, path.WithKey(key))
				}
			}
		}
	}

	if len(inputs) < 1 {
		return
	}

	var prog *utils.ProgressBar
	if !args.NoProg {
		prog, _ = utils.NewProgressBar(
			pterm.DefaultProgressbar.WithTitle("Analyzing paths").
				WithTotal(len(inputs)).WithRemoveWhenDone(true),
		).Start()
	}

	var filesHot, filesWarm, filesCold,
		lastModified,
		totalSize, totalDirs, totalFiles int64

	slot := utils.NewSlot(args.Concurrency)
	for stats := range obsAnalyzer(slot, prog, args.ObsClient, inputs, args.MinChunks) {
		if stats.Exists {
			utils.WriteOutput(stats)
			pterm.Printf(
				"obs://%s/%s: size: %d (%s), objects: %d (%d dirs, %d files [%d/%d/%d]), last modified: %s\n",
				stats.Bucket, stats.Key,
				stats.Size, utils.FormatSize(stats.Size),
				stats.DirCount+stats.FileCount,
				stats.DirCount, stats.FileCount,
				stats.Fsc.Hot, stats.Fsc.Warm, stats.Fsc.Cold,
				time.Unix(0, stats.LastModified*int64(time.Millisecond)).Format("2006-01-02 15:04:05.000"),
			)

			if args.Summarize {
				filesHot += stats.Fsc.Hot
				filesWarm += stats.Fsc.Warm
				filesCold += stats.Fsc.Cold
				totalSize += stats.Size
				totalDirs += stats.DirCount
				totalFiles += stats.FileCount
				lastModified = max(lastModified, stats.LastModified)

				if stats.IsDir() {
					totalDirs++
				}
			}
		} else {
			logger.Error(stats.URI() + ": no such file or directory")
		}
	}
	slot.Close()

	if prog != nil {
		prog.Stop()
	}

	if args.Summarize {
		pterm.Println()
		pterm.Printf(
			"Total size: %d (%s), objects: %d (%d dirs, %d files [%d/%d/%d]), last modified: %s\n",
			totalSize, utils.FormatSize(totalSize),
			totalDirs+totalFiles,
			totalDirs, totalFiles,
			filesHot, filesWarm, filesCold,
			time.Unix(0, lastModified*int64(time.Millisecond)).Format("2006-01-02 15:04:05.000"),
		)
	}
}

func obsAnalyzer(
	slot *utils.Slot, prog *utils.ProgressBar,
	obsClient *obs.ObsClient, inputs []obs.ObsPath, minChunks int,
) <-chan obs.ObsPathAnalyzed {
	var (
		stats   = make(chan obs.ObsPathAnalyzed, 1)
		subSlot = utils.NewSlot(slot.Concurrency)
	)
	if minChunks > 1 {
		slot.DoNow(func() {
			for _, p := range inputs {
				c := make(chan obs.ObsPathChunked, 1)
				slot.Do(func() {
					chunk := obsClient.SplitChunk(minChunks, p)
					logger.Debug("path chunked.", logger.Args(
						"countDirs", len(chunk.Dirs),
						"countFiles", len(chunk.Files),
						"extraDirs", chunk.ExtraDirs,
					))
					c <- chunk
					close(c)
				})
				subSlot.Do(func() {
					s := obsClient.AnalyzeChunk(slot, <-c)
					if prog != nil {
						prog.Increment()
					}
					stats <- s
				})
			}
			subSlot.Close()
			close(stats)
		})
	} else {
		slot.DoNow(func() {
			for _, p := range inputs {
				subSlot.Do(func() {
					s := obsClient.Analyze(p)
					if prog != nil {
						prog.Increment()
					}
					stats <- s
				})
			}
			subSlot.Close()
			close(stats)
		})
	}
	return stats
}
