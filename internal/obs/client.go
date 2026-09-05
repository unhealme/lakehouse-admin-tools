package obs

import (
	"iter"
	"slices"
	"strings"
	"sync"

	"github.com/gobwas/glob"
	"github.com/huaweicloud/huaweicloud-sdk-go-obs/obs"
	"github.com/pterm/pterm"
	"github.com/unhealme/lakehouse-admin-tools/utils"
)

type ObsClient struct{ *obs.ObsClient }

func (c ObsClient) iterPaths(logger *pterm.Logger, input obs.ListObjectsInput, depth int, dirOnly bool) iter.Seq[ObsPathContent] {
	return func(yield func(ObsPathContent) bool) {
		obsPathStr := "obs://" + input.Bucket + "/" + input.Prefix
		for p := 1; true; p++ {
			logArgs := logger.Args("depth", depth, "path", obsPathStr, "page", p)
			logger.Debug("listing obs paths.", logArgs)
			r, err := c.ListObjects(&input)
			if err != nil {
				// obsError, ok := err.(obs.ObsError)
				logger.Error("unable to list obs paths.", logArgs, logger.Args("error", err))
				break
			}
			logger.Debug("obs paths fetched.", logArgs, logger.Args("contents", len(r.Contents), "common-prefix", len(r.CommonPrefixes)))
			for _, v := range r.CommonPrefixes {
				if !yield(NewObsPathContent(depth, input.Bucket, v, nil)) {
					return
				}
			}
			if !dirOnly {
				for _, v := range r.Contents {
					if !yield(NewObsPathContent(depth, input.Bucket, v.Key, &v)) {
						return
					}
				}
			}
			if !r.IsTruncated {
				break
			}
			input.Marker = r.NextMarker
		}
	}
}

func (c ObsClient) Analyze(logger *pterm.Logger, slot *utils.Slot, minChunks int, path ObsPath) ObsPathAnalyzed {
	var (
		once sync.Once

		dirs, files, extraDirs = c.SplitChunk(logger, minChunks, path)
		stats                  = ObsPathAnalyzed{
			Bucket:   path.Bucket,
			Key:      path.Key,
			DirCount: int64(extraDirs),
		}
	)
	logger.Debug("path chunked.", logger.Args(
		"countDirs", len(dirs),
		"countFiles", len(files),
		"extraDirs", extraDirs,
	))
	for _, f := range files {
		stats.FileCount++
		stats.Size += f.Content.Size
		if f.Content.LastModified.UnixMilli() > stats.LastModified {
			stats.LastModified = f.Content.LastModified.UnixMilli()
		}
	}

	if slot == nil {
		slot = utils.NewSlot(0)
	}

	slot.Unblock()
	for paths := range slot.MapValue(func(p ObsPath) []ObsPathContent {
		return slices.Collect(c.Walk0(logger, p, false))
	}, dirs, false) {
		for _, op := range paths {
			suffix := strings.TrimPrefix(op.Key, strings.TrimSuffix(path.Key, "/"))
			if suffix != "" && !strings.HasPrefix(suffix, "/") {
				continue
			}

			once.Do(func() {
				if suffix != "" && !strings.HasSuffix(stats.Key, "/") {
					stats.Key += "/"
				}
				stats.Exists = true
			})

			if op.IsDir() {
				if suffix != "/" {
					stats.DirCount++
				}
			} else {
				stats.FileCount++
				stats.Size += op.Content.Size
				if op.Content.LastModified.UnixMilli() > stats.LastModified {
					stats.LastModified = op.Content.LastModified.UnixMilli()
				}
			}
		}
	}
	if stats.DirCount > 0 || stats.FileCount > 0 {
		stats.Exists = true
	}
	slot.Block()
	return stats
}

func (c ObsClient) Glob(logger *pterm.Logger, path ObsPath) (matchKeys []string) {
	if _, err := glob.Compile(path.Key, '/'); err != nil {
		return
	}
	splitKeys := splitGlobSegments(path.Key)
	for _, key := range splitKeys {
		if !key.isGlob {
			if len(matchKeys) < 1 {
				matchKeys = append(matchKeys, key.segment)
			} else {
				for i, k := range matchKeys {
					matchKeys[i] = k + key.segment
				}
			}
		} else {
			if len(matchKeys) < 1 {
				matchKeys = append(matchKeys, "/")
			}
			var nextKeys []string
			g := glob.MustCompile(key.segment)
			for _, k := range matchKeys {
				if !strings.HasSuffix(k, "/") {
					k += "/"
				}
				for op := range c.Walk(logger, NewObsPath(path.Bucket, k), 1, false) {
					if name := strings.TrimPrefix(op.Key, k); name != "" && g.Match(name) {
						nextKeys = append(nextKeys, op.Key)
					}
				}
			}
			matchKeys = nextKeys
		}
	}
	return
}

func (c ObsClient) SplitChunk(logger *pterm.Logger, minChunks int, path ObsPath) (
	dirs []ObsPath, files []ObsPathContent, extraDirs int,
) {
	if minChunks <= 1 {
		return []ObsPath{path}, nil, 0
	}

	i := obs.ListObjectsInput{
		Bucket:       path.Bucket,
		MaxKeys:      1000,
		EncodingType: "url",
		Delimiter:    "/",
	}

	nextDepth := []string{path.Key}
	for depth := 1; len(nextDepth) > 0; depth++ {
		paths := nextDepth[:]
		nextDepth = nil
		for n, p := range paths {
			i.Prefix = p
			for op := range c.iterPaths(logger, i, depth, false) {
				if op.IsDir() && !SameObsKey(p, op.Key) {
					nextDepth = append(nextDepth, op.Key)
				}
				if !op.IsDir() {
					files = append(files, op)
				}
			}
			if len(dirs)+len(files)+len(nextDepth)+len(paths)-n+1 >= minChunks {
				for _, dir := range nextDepth {
					dirs = append(dirs, NewObsPath(path.Bucket, dir))
				}
				nextDepth = nil
				paths = paths[n+1:]
				extraDirs--
				break
			}
		}
		if len(nextDepth) < 1 {
			for _, dir := range paths {
				dirs = append(dirs, NewObsPath(path.Bucket, dir))
			}
		}
		extraDirs++
	}
	return
}

func (c ObsClient) Walk(logger *pterm.Logger, path ObsPath, maxDepth int, dirOnly bool) iter.Seq[ObsPathContent] {
	i := obs.ListObjectsInput{
		Bucket:       path.Bucket,
		MaxKeys:      1000,
		EncodingType: "url",
		Delimiter:    "/",
	}

	return func(yield func(ObsPathContent) bool) {
		nextDepth := []string{path.Key}
		for depth := 1; (depth <= maxDepth || maxDepth < 0) && len(nextDepth) > 0; depth++ {
			dirs := nextDepth[:]
			nextDepth = nil
			for _, p := range dirs {
				i.Prefix = p
				for op := range c.iterPaths(logger, i, depth, dirOnly) {
					if depth == 1 || !SameObsKey(p, op.Key) {
						if !yield(op) {
							return
						}
					}
					if op.IsDir() && !SameObsKey(p, op.Key) {
						nextDepth = append(nextDepth, op.Key)
					}
				}
			}
		}
	}
}

func (c ObsClient) Walk0(logger *pterm.Logger, path ObsPath, dirOnly bool) iter.Seq[ObsPathContent] {
	i := obs.ListObjectsInput{
		Bucket:       path.Bucket,
		MaxKeys:      1000,
		EncodingType: "url",
		Prefix:       path.Key,
	}
	return c.iterPaths(logger, i, -1, dirOnly)
}

func (c ObsClient) RenameObject(logger *pterm.Logger, path ObsPath, keyAfter string) {
	fullKey := path.URI()
	argsOk := logger.Args("before", fullKey, "after", path.WithKey(keyAfter).URI())
	if strings.HasSuffix(path.Key, "/") {
		_, err := c.RenameFolder(&obs.RenameFolderInput{
			Bucket:       path.Bucket,
			Key:          path.Key,
			NewObjectKey: keyAfter,
		})
		if err != nil {
			logger.Warn("unable to rename directory.", logger.Args("dir", fullKey, "error", err))
		} else {
			logger.Debug("rename directory success.", argsOk)
		}
	} else {
		_, err := c.RenameFile(&obs.RenameFileInput{
			Bucket:       path.Bucket,
			Key:          path.Key,
			NewObjectKey: keyAfter,
		})
		if err != nil {
			logger.Warn("unable to rename file.", logger.Args("file", fullKey, "error", err))
		} else {
			logger.Debug("rename file success.", argsOk)
		}
	}
}

func (c ObsClient) SetStorageClass(logger *pterm.Logger, path ObsPath, class obs.StorageClassType) {
	_, err := c.SetObjectMetadata(&obs.SetObjectMetadataInput{
		Bucket:            path.Bucket,
		Key:               path.Key,
		MetadataDirective: obs.ReplaceNew,
		StorageClass:      class,
	})
	if err != nil {
		logger.Warn("unable to set storage class for object.", logger.Args("path", path.URI(), "error", err))
	} else {
		logger.Debug("set storage class for object success.", logger.Args("path", path.URI(), "class", class))
	}
}

func NewClient(endpoint string, ak, sk string, token string) (*ObsClient, error) {
	base, err := obs.New(
		ak, sk, endpoint,
		obs.WithSecurityToken(token),
		obs.WithSecurityProviders(obs.NewEcsSecurityProvider(1)),
		obs.WithProxyFromEnv(true),
	)
	if err != nil {
		return nil, err
	}
	return &ObsClient{base}, nil
}
