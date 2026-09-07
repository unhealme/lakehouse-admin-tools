package commands

import (
	"context"
	"iter"
	rand "math/rand/v2"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"time"

	yaml "github.com/goccy/go-yaml"
	"github.com/pterm/pterm"
	"github.com/unhealme/lakehouse-admin-tools/internal"
	"github.com/unhealme/lakehouse-admin-tools/internal/clients/obs"
	"github.com/unhealme/lakehouse-admin-tools/pkg/arguments"
	"github.com/unhealme/lakehouse-admin-tools/pkg/utils"
)

const ObsBatchSetStorageClassVersion = "2026.09.05-0"

type ObsBatchSetStorageClassInput struct {
	Path        string
	DateRange   utils.DateRangeParsed `yaml:"date-range"`
	TargetClass obs.StorageClassType  `yaml:"target-class"`
	Exclude     []string
}

func ObsBatchSetStorageClass(logger *pterm.Logger, args *arguments.ObsBatchSetStorageClassArgs) {
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

func processBatchSetStorageClassInput(logger *pterm.Logger, input ObsBatchSetStorageClassInput, args *arguments.ObsBatchSetStorageClassArgs) {
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
	if dR.Kind != utils.DateRangeArray {
		excludes := utils.SliceToSet(input.Exclude)
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
	case utils.DateRangeConstraint:
		for par := range parents {
			parsed, err := utils.ParseStrftime(par.Name(), dR.Format)
			if err != nil {
				logger.Warn("unable to parse path date. skipping..", logger.Args("path", par.Key, "format", dR.Format, "error", err))
				continue
			}
			if (dR.End == nil || !parsed.After(*dR.End)) && (dR.Start == nil || !parsed.Before(*dR.Start)) {
				actualRun(par.Key)
			}
		}
	case utils.DateRangePattern:
		for par := range parents {
			if match, _ := filepath.Match(dR.Pattern, par.Name()); match {
				actualRun(par.Key)
			}
		}
	case utils.DateRangeRegex:
		re, err := regexp.Compile(dR.Regex)
		if err != nil {
			logger.Fatal("unable to compile regex pattern.", logger.Args("pattern", dR.Regex, "error", err))
		}
		for par := range parents {
			if re.MatchString(par.Name()) {
				actualRun(par.Key)
			}
		}
	case utils.DateRangeArray:
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
	slot := utils.NewSlot(max(concurrency, 1))
	if noProg {
		slot.Map(
			func(path obs.ObsPathContent) {
				if !path.IsDir() {
					obsClient.SetStorageClass(logger, basePath.WithKey(path.Key), storageClass)
				}
			},
			slices.Collect(walker),
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
		ctx, done := context.WithCancel(context.Background())
		prog, _ := utils.NewProgressBar(ctx).WithTitle("Setting Storage Class").WithTotal(total).Start()
		defer done()
		slot.Map(
			func(key string) {
				obsClient.SetStorageClass(logger, basePath.WithKey(key), storageClass)
				prog.Increment()
			},
			keys,
		)
	}
}
