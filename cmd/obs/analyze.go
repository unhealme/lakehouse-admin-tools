package obs

import (
	"fmt"
	"strings"
	"sync"

	"github.com/unhealme/lakehouse-admin-tools/cmd"
	"github.com/unhealme/lakehouse-admin-tools/internal"
	"github.com/unhealme/lakehouse-admin-tools/internal/obs"
)

func Analyze(logger *internal.Logger, args *cmd.ObsAnalyzeArgs) {
	logger.Debug("using analyze args.", logger.Args(internal.ToArgs(*args)...))
	for _, uri := range args.Paths {
		inputPath, err := obs.PathFromURI(uri)
		if err != nil {
			logger.Warn("skipping path due to error.", logger.Args("path", uri, "error", err))
			continue
		}
		var (
			files int
			dirs  int
			size  int64

			inputIsFile bool
			once        sync.Once
		)
		for op := range args.ObsClient.Walk0(logger, inputPath.Bucket, inputPath.Key, args.DirOnly) {
			suffix, _ := strings.CutPrefix(op.Key, strings.TrimRight(inputPath.Key, "/"))
			once.Do(func() { inputIsFile = !strings.HasPrefix(suffix, "/") })
			if inputIsFile && suffix != "" {
				continue
			}
			if !inputIsFile && suffix == "/" {
				continue
			}

			if op.IsDir() {
				dirs++
			} else {
				files++
				size += op.Content.Size
			}
		}
		if dirs+files < 1 {
			fmt.Printf("%s: no such file or directory", uri)
		} else {
			fmt.Printf(
				"%s: size: %d (%s), objects: %d (%d dirs, %d files)\n",
				uri, size, internal.FormatSize(size), dirs+files, dirs, files,
			)
		}
	}
}
