package obs

import (
	"strings"
	"sync"

	"github.com/unhealme/lakehouse-admin-tools/internal"
	obs "github.com/unhealme/lakehouse-admin-tools/internal/obs"
)

const BatchRenameVersion = "2026.04.09-3"

var logger = internal.DefaultLogger()

type BatchRenameArgs struct {
	Bucket string `arg:"positional,required" placeholder:"BUCKET"`
	Path   string `arg:"positional,required" placeholder:"PATH"`
	Prefix string `arg:"positional,required" placeholder:"PREFIX"`

	Concurrency int  `arg:"-j,--" default:"2" help:"max rename job concurrency" placeholder:"NUM"`
	DirOnly     bool `arg:"-d,--dir-only" help:"only rename directories"`
	DryRun      bool `arg:"-,--dry-run"`
}

func BatchRename(obsClient *obs.ObsClient, args *BatchRenameArgs) {
	var (
		paths [][2]string
		total = 0
	)
	for p := range obsClient.IterPaths(args.Bucket, args.Path, args.DirOnly) {
		if name, _ := strings.CutPrefix(p, args.Path); len(name) > 0 && !strings.HasPrefix(name, args.Prefix) {
			total++
			after := args.Path + args.Prefix + name
			paths = append(paths, [2]string{p, after})
		}
	}
	logger.Info("all obs paths fetched.", logger.Args("path to rename", total))

	if total > 0 {
		prog, _ := internal.NewProgressBar().WithTitle("Renaming paths").WithTotal(total).Start()
		defer prog.Stop()
		sem := make(chan struct{}, args.Concurrency)
		var wg sync.WaitGroup
		for _, p := range paths {
			sem <- struct{}{}
			prog.Increment()
			wg.Go(func() {
				if !args.DryRun {
					obsClient.RenameObject(args.Bucket, p[0], p[1])
				} else {
					logger.Info("renaming directory.", logger.Args("before", p[0], "after", p[1]))
				}
				<-sem
			})
		}
		wg.Wait()
	}
	logger.Info("rename obs paths done.")
}
