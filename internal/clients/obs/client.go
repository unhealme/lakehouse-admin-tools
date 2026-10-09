package obs

import (
	"fmt"
	"io"
	"iter"
	"strings"

	"github.com/gobwas/glob"
	"github.com/huaweicloud/huaweicloud-sdk-go-obs/obs"
	"github.com/unhealme/lakehouse-admin-tools/internal/logger"
	"github.com/unhealme/lakehouse-admin-tools/pkg/utils"
)

type ObsClient struct{ *obs.ObsClient }

func (c ObsClient) Analyze(path ObsPath) ObsPathAnalyzed {
	stats := ObsPathAnalyzed{ObsPath: path}
	for op := range c.Walk0(path, false) {
		suffix := strings.TrimPrefix(op.Key, strings.TrimSuffix(path.Key, "/"))
		if suffix != "" && !strings.HasPrefix(suffix, "/") {
			// skip prefix only match
			// prefix => prefix-path
			continue
		}

		switch {
		case suffix == "/":
			stats.Exists = true
		case op.IsDir():
			stats.DirCount++
		default:
			stats.FileCount++
			stats.Size += op.Content.Size
			stats.LastModified = max(
				stats.LastModified, op.Content.LastModified.UnixMilli(),
			)
			switch string(op.Content.StorageClass) {
			case "STANDARD":
				stats.Fsc.Hot++
			case "WARM":
				stats.Fsc.Warm++
			case "COLD", "DEEP_ARCHIVE", "GLACIER":
				stats.Fsc.Cold++
			}
		}
	}
	if stats.FileCount > 0 || stats.DirCount > 0 {
		stats.Exists = true
	}
	return stats
}

func (c ObsClient) AnalyzeChunk(slot *utils.Slot, chunk ObsPathChunked) ObsPathAnalyzed {
	stats := ObsPathAnalyzed{
		ObsPath:  chunk.ObsPath,
		DirCount: int64(len(chunk.Dirs) + chunk.ExtraDirs - 1),
		Exists:   chunk.Count() > 0,
	}
	for _, f := range chunk.Files {
		stats.FileCount++
		stats.Size += f.Content.Size
		stats.LastModified = max(
			stats.LastModified, f.Content.LastModified.UnixMilli(),
		)
		switch string(f.Content.StorageClass) {
		case "STANDARD":
			stats.Fsc.Hot++
		case "WARM":
			stats.Fsc.Warm++
		case "COLD", "DEEP_ARCHIVE", "GLACIER":
			stats.Fsc.Cold++
		}
	}

	for ds := range slot.MapValue(func(path ObsPath) ObsPathAnalyzed {
		return c.Analyze(path)
	}, chunk.Dirs, false) {
		stats.DirCount += ds.DirCount
		stats.FileCount += ds.FileCount
		stats.Size += ds.Size

		stats.LastModified = max(stats.LastModified, ds.LastModified)

		stats.Fsc.Hot += ds.Fsc.Hot
		stats.Fsc.Warm += ds.Fsc.Warm
		stats.Fsc.Cold += ds.Fsc.Cold
	}
	return stats
}

func (c ObsClient) Glob(path ObsPath) (matchKeys []string) {
	if _, err := glob.Compile(path.Key, '/'); err != nil {
		return
	}
	splitKeys := splitGlobSegments(path.Key)
	for _, key := range splitKeys {
		if !key.isGlob {
			if len(matchKeys) < 1 {
				matchKeys = append(matchKeys, key.segment)
			} else {
				var nextKeys []string
				for _, k := range matchKeys {
					for op := range c.Walk(path.WithKey(k), 1, false) {
						if SameObsKey(strings.TrimPrefix(op.Key, k), key.segment) {
							logger.Debug(fmt.Sprintf("%s match with %s", op.Key, key.segment))
							nextKeys = append(nextKeys, op.Key)
						}
					}
				}
				if len(nextKeys) < 1 {
					return nil
				}
				matchKeys = nextKeys
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
				for op := range c.Walk(path.WithKey(k), 1, false) {
					name := strings.TrimPrefix(op.Key, k)
					if name != "" && g.Match(name) {
						nextKeys = append(nextKeys, op.Key)
					}
				}
			}
			if len(nextKeys) < 1 {
				return nil
			}
			matchKeys = nextKeys
		}
	}
	return
}

func (c ObsClient) ReadFile(path ObsPath) (io.ReadCloser, error) {
	i := obs.GetObjectInput{Bucket: path.Bucket, Key: path.Key}
	resp, err := c.GetObject(&i)
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}

func (c ObsClient) RenameObject(path ObsPath, keyAfter string) {
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

func (c ObsClient) SetStorageClass(path ObsPath, class obs.StorageClassType) {
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

func (c ObsClient) SplitChunk(minChunks int, path ObsPath) (chunk ObsPathChunked) {
	chunk = ObsPathChunked{ObsPath: path}
	i := obs.ListObjectsInput{
		Bucket:       path.Bucket,
		MaxKeys:      1000,
		EncodingType: "url",
		Delimiter:    "/",
	}

	var paths []string
	nextDepth := []string{path.Key}
	for depth := 1; len(nextDepth) > 0; depth++ {
		paths, nextDepth = nextDepth, nil
		for n, p := range paths {
			i.Prefix = p
			for op := range c.iterPaths(i, depth, false) {
				suffix := strings.TrimPrefix(op.Key, strings.TrimSuffix(p, "/"))
				if suffix != "" && !strings.HasPrefix(suffix, "/") {
					// skip prefix only match
					// prefix => prefix-path
					continue
				}

				switch {
				case op.Key == p:
					chunk.ExtraDirs++
				case op.IsDir():
					nextDepth = append(nextDepth, op.Key)
				default:
					chunk.Files = append(chunk.Files, op)
				}
			}
			if len(chunk.Dirs)+len(nextDepth)+len(paths[n+1:]) >= minChunks { // chunk is full
				for _, dir := range append(nextDepth, paths[n+1:]...) {
					chunk.Dirs = append(chunk.Dirs, path.WithKey(dir))
				}
				return
			}
		}
	}
	return
}

func (c ObsClient) Walk(path ObsPath, maxDepth int, dirOnly bool) iter.Seq[ObsPathContent] {
	i := obs.ListObjectsInput{
		Bucket:       path.Bucket,
		MaxKeys:      1000,
		EncodingType: "url",
		Delimiter:    "/",
	}

	return func(yield func(ObsPathContent) bool) {
		var dirs, next []string
		next = []string{path.Key}
		for depth := 1; (depth <= maxDepth || maxDepth < 0) && len(next) > 0; depth++ {
			dirs, next = next, nil
			for _, p := range dirs {
				i.Prefix = p
				for op := range c.iterPaths(i, depth, dirOnly) {
					if depth == 1 || !SameObsKey(p, op.Key) {
						if !yield(op) {
							return
						}
					}
					if op.IsDir() && !SameObsKey(p, op.Key) {
						next = append(next, op.Key)
					}
				}
			}
		}
	}
}

func (c ObsClient) Walk0(path ObsPath, dirOnly bool) iter.Seq[ObsPathContent] {
	i := obs.ListObjectsInput{
		Bucket:       path.Bucket,
		MaxKeys:      1000,
		EncodingType: "url",
		Prefix:       path.Key,
	}
	return c.iterPaths(i, -1, dirOnly)
}

func (c ObsClient) WriteFile(path ObsPath, data io.Reader) error {
	i := obs.PutObjectInput{Bucket: path.Bucket, Key: path.Key, Body: data}
	if _, err := c.PutObject(&i); err != nil {
		return err
	}
	return nil
}

func (c ObsClient) iterPaths(i obs.ListObjectsInput, depth int, dirOnly bool) iter.Seq[ObsPathContent] {
	return func(yield func(ObsPathContent) bool) {
		path := "obs://" + i.Bucket + "/" + i.Prefix
		for p := 1; true; p++ {
			logArgs := logger.Args("path", path, "page", p)
			if depth > 0 {
				logArgs = append(logger.Args("depth", depth), logArgs...)
			}
			logger.Debug("listing obs paths.", logArgs)
			r, err := c.ListObjects(&i)
			if err != nil {
				// obsError, ok := err.(obs.ObsError)
				logger.Error("unable to list obs paths.", logArgs, logger.Args("error", err))
				break
			}
			logger.Debug("obs paths fetched.", logArgs, logger.Args("contents", len(r.Contents), "common-prefix", len(r.CommonPrefixes)))
			for _, v := range r.CommonPrefixes {
				if !yield(NewObsPathContent(depth, i.Bucket, v, nil)) {
					return
				}
			}
			if !dirOnly {
				for _, v := range r.Contents {
					if !yield(NewObsPathContent(depth, i.Bucket, v.Key, &v)) {
						return
					}
				}
			}
			if !r.IsTruncated {
				break
			}
			i.Marker = r.NextMarker
		}
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
