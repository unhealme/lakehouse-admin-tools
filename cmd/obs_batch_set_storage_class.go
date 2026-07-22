package cmd

import (
	"iter"
	rand "math/rand/v2"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/goccy/go-yaml"
	"github.com/pterm/pterm"
	"github.com/unhealme/lakehouse-admin-tools/internal"
	"github.com/unhealme/lakehouse-admin-tools/internal/obs"
)

const ObsBatchSetStorageClassVersion = "2026.06.22-0"

type ObsBatchSetStorageClassInput struct {
	Path        string
	DateRange   internal.DateRangeParsed `yaml:"date-range"`
	TargetClass obs.StorageClassType     `yaml:"target-class"`
	Exclude     []string
}

func ObsBatchSetStorageClass(logger *pterm.Logger, args *ObsBatchSetStorageClassArgs) {
	logger.Debug("using batch set storage class args.", logger.Args(internal.ToArgs(*args)...))
	for _, inputFile := range args.InputFiles {
		buf, err := os.ReadFile(inputFile)
		if err != nil {
			logger.Warn("unable to read input file. skipping..", logger.Args("file", inputFile, "error", err))
			continue
		}
		var inputs []ObsBatchSetStorageClassInput
		if err := yaml.Unmarshal(buf, &inputs); err != nil {
			logger.Warn("unable to parse input file. skipping..", logger.Args("file", inputFile, "error", err))
			continue
		}
		for _, input := range inputs {
			processBatchSetStorageClassInput(logger, input, args)
		}
	}
}

func processBatchSetStorageClassInput(logger *pterm.Logger, input ObsBatchSetStorageClassInput, args *ObsBatchSetStorageClassArgs) {
	inputPath, err := obs.PathFromURI(input.Path)
	if err != nil {
		logger.Warn("skipping input due to error.", logger.Args("path", input.Path, "error", err))
		return
	}
	actualRun := func(key string) {
		if !args.DryRun {
			processSetStorageClass(logger, args.ObsClient, inputPath.WithKey(key), input.TargetClass, args.NoProg, args.Concurrency)
		} else {
			logger.Info("setting storage class for object.", logger.Args("path", inputPath.WithKey(key).URI(), "class", input.TargetClass))
			time.Sleep(200 + rand.N(300*time.Millisecond))
		}
	}

	var parents iter.Seq[obs.ObsPathContent]
	dR := input.DateRange
	if dR.Kind != internal.DateRangeArray {
		excludes := internal.SliceToSet(input.Exclude)
		parents = func(yield func(obs.ObsPathContent) bool) {
			if !strings.HasSuffix(inputPath.Key, "/") {
				inputPath.Key += "/"
			}
			for p := range args.ObsClient.Walk(logger, *inputPath, 1, true) {
				if _, skip := excludes[p.Name()]; !skip {
					if !yield(p) {
						return
					}
				}
			}
		}
	}

	switch dR.Kind {
	case internal.DateRangeConstraint:
		for par := range parents {
			var parsed time.Time
			if err := internal.ParseStrftime(par.Name(), dR.Format, &parsed); err != nil {
				logger.Warn("unable to parse path date. skipping..", logger.Args("path", par.Key, "format", dR.Format, "error", err))
				continue
			}
			if (dR.End == nil || !parsed.After(*dR.End)) && (dR.Start == nil || !parsed.Before(*dR.Start)) {
				actualRun(par.Key)
			}
		}
	case internal.DateRangePattern:
		for par := range parents {
			if match, _ := filepath.Match(dR.Pattern, par.Name()); match {
				actualRun(par.Key)
			}
		}
	case internal.DateRangeRegex:
		re, err := regexp.Compile(dR.Regex)
		if err != nil {
			logger.Fatal("unable to compile regex pattern.", logger.Args("pattern", dR.Regex, "error", err))
		}
		for par := range parents {
			if re.MatchString(par.Name()) {
				actualRun(par.Key)
			}
		}
	case internal.DateRangeArray:
		for _, base := range dR.Array {
			actualRun(path.Join(inputPath.Key, base))
		}
	}
}

func processSetStorageClass(logger *pterm.Logger, obsClient *obs.ObsClient, basePath obs.ObsPath, storageClass obs.StorageClassType, noProg bool, concurrency int) {
	if !strings.HasSuffix(basePath.Key, "/") {
		basePath.Key += "/"
	}
	walker := obsClient.Walk(logger, basePath, -1, false)
	if noProg {
		internal.ParallelMap(
			func(path obs.ObsPathContent) {
				if !path.IsDir() {
					obsClient.SetStorageClass(logger, basePath.WithKey(path.Key), storageClass)
				}
			},
			walker,
			concurrency,
		)
	} else {
		var keys []string
		var total int
		for path := range walker {
			if !path.IsDir() {
				keys = append(keys, path.Key)
				total += 1
			}
		}
		prog, _ := internal.NewProgressBar().WithTitle("Setting Storage Class").WithTotal(total).Start()
		defer prog.Stop()
		internal.ParallelMap(
			func(key string) {
				obsClient.SetStorageClass(logger, basePath.WithKey(key), storageClass)
				prog.Increment()
			},
			slices.Values(keys),
			concurrency,
		)
	}
}
