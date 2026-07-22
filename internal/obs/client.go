package obs

import (
	"fmt"
	"iter"
	"strings"
	"sync"

	"github.com/gobwas/glob"
	"github.com/huaweicloud/huaweicloud-sdk-go-obs/obs"
	"github.com/pterm/pterm"
)

const GlobToken = "*?\\[]{}"

type ObsClient struct{ *obs.ObsClient }

func (c ObsClient) iterPaths(logger *pterm.Logger, input obs.ListObjectsInput, depth int, dirOnly bool) iter.Seq[ObsPathContent] {
	return func(yield func(ObsPathContent) bool) {
		p := 1
		for {
			logArgs := logger.Args("depth", depth, "path", fmt.Sprintf("obs://%s/%s", input.Bucket, input.Prefix), "page", p)
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
			p++
		}
	}
}

func (c ObsClient) Analyze(logger *pterm.Logger, path ObsPath) ObsPathAnalyzed {
	var (
		stats ObsPathAnalyzed

		inputIsFile bool
		once        sync.Once
	)
	stats.Bucket = path.Bucket
	stats.Key = path.Key
	for op := range c.Walk0(logger, path, false) {
		suffix := strings.TrimPrefix(op.Key, strings.TrimSuffix(path.Key, "/"))
		once.Do(func() { inputIsFile = !strings.HasPrefix(suffix, "/") })
		if inputIsFile && suffix != "" {
			continue
		}
		if !inputIsFile && suffix == "/" {
			stats.Exists = true
			if !strings.HasSuffix(stats.Key, "/") {
				stats.Key += "/"
			}
			continue
		}
		if op.IsDir() {
			stats.DirCount++
		} else {
			stats.FileCount++
			stats.Size += op.Content.Size
		}
	}
	return stats
}

func (c ObsClient) Glob(logger *pterm.Logger, path ObsPath) (matchKeys []string) {
	if _, err := glob.Compile(path.Key, '/'); err != nil {
		return
	}

	var (
		keySegment string
		splitKeys  []string
	)
	for segment := range strings.SplitSeq(path.Key, "/") {
		if !strings.ContainsAny(segment, GlobToken) {
			keySegment += "/" + segment
		} else {
			if keySegment != "" {
				splitKeys = append(splitKeys, keySegment)
				keySegment = ""
			}
			splitKeys = append(splitKeys, segment)
		}
	}
	if keySegment != "" {
		splitKeys = append(splitKeys, keySegment)
	}

	for _, key := range splitKeys {
		key = strings.TrimPrefix(key, "/")
		if !strings.ContainsAny(key, GlobToken) {
			if len(matchKeys) < 1 {
				matchKeys = append(matchKeys, key)
			} else {
				for i, k := range matchKeys {
					matchKeys[i] = k + key
				}
			}
		} else {
			if len(matchKeys) < 1 {
				matchKeys = append(matchKeys, "/")
			}
			var nextKeys []string
			g := glob.MustCompile(key)
			for _, k := range matchKeys {
				if !strings.HasSuffix(k, "/") {
					k += "/"
				}
				for op := range c.Walk(logger, NewObsPath(path.Bucket, k), 1, false) {
					if name := strings.TrimPrefix(op.Key, k); name != "" && g.Match(name) {
						nextKeys = append(nextKeys, k+name)
					}
				}
			}
			matchKeys = nextKeys
		}
	}
	return
}

func (c ObsClient) Walk(logger *pterm.Logger, path ObsPath, maxDepth int, dirOnly bool) iter.Seq[ObsPathContent] {
	i := obs.ListObjectsInput{}
	i.Bucket = path.Bucket
	i.MaxKeys = 1000
	i.EncodingType = "url"
	i.Delimiter = "/"
	return func(yield func(ObsPathContent) bool) {
		nextDepth := []string{path.Key}
		for depth := 1; (depth <= maxDepth || maxDepth < 0) && len(nextDepth) > 0; depth++ {
			dirs := nextDepth[:]
			nextDepth = nil
			for _, p := range dirs {
				i.Prefix = p
				for op := range c.iterPaths(logger, i, depth, dirOnly) {
					if !yield(op) {
						return
					}
					if op.IsDir() {
						nextDepth = append(nextDepth, op.Key)
					}
				}
			}
		}
	}
}

func (c ObsClient) Walk0(logger *pterm.Logger, path ObsPath, dirOnly bool) iter.Seq[ObsPathContent] {
	i := obs.ListObjectsInput{}
	i.Bucket = path.Bucket
	i.MaxKeys = 1000
	i.EncodingType = "url"
	i.Prefix = path.Key
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
	base, err := obs.New(ak, sk, endpoint,
		obs.WithSecurityToken(token),
		obs.WithSecurityProviders(obs.NewEcsSecurityProvider(1)),
		obs.WithProxyFromEnv(true),
	)
	if err != nil {
		return nil, err
	}
	return &ObsClient{base}, nil
}
