package obs

import (
	"fmt"
	"iter"
	"strings"

	"github.com/huaweicloud/huaweicloud-sdk-go-obs/obs"
	"github.com/pterm/pterm"
)

type ObsClient struct{ *obs.ObsClient }

func (c ObsClient) iterPaths(logger *pterm.Logger, input obs.ListObjectsInput, depth int, dirOnly bool) iter.Seq[ObsPath] {
	return func(yield func(ObsPath) bool) {
		p := 1
		for {
			logArgs := logger.Args("depth", depth, "path", fmt.Sprintf("obs://%s/%s", input.Bucket, input.Prefix), "page", p)
			logger.Debug("listing obs paths.", logArgs)
			r, err := c.ListObjects(&input)
			if err != nil {
				// obsError, ok := err.(obs.ObsError)
				logger.Error("unable to list obs paths.", append(logArgs, logger.Args("error", err)...))
				break
			}
			logger.Debug("obs paths fetched.", append(logArgs, logger.Args("contents", len(r.Contents), "common-prefix", len(r.CommonPrefixes))...))
			for _, v := range r.CommonPrefixes {
				if !yield(NewPath(depth, input.Bucket, v, nil)) {
					return
				}
			}
			if !dirOnly {
				for _, v := range r.Contents {
					if !yield(NewPath(depth, input.Bucket, v.Key, &v)) {
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

func (c ObsClient) Walk(logger *pterm.Logger, bucket, basePath string, maxDepth int, dirOnly bool) iter.Seq[ObsPath] {
	i := obs.ListObjectsInput{}
	i.Bucket = bucket
	i.MaxKeys = 1000
	i.EncodingType = "url"
	i.Delimiter = "/"
	return func(yield func(ObsPath) bool) {
		nextDepth := []string{basePath}
		for depth := 1; depth <= maxDepth || maxDepth < 0; depth++ {
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
			if len(nextDepth) < 1 {
				break
			}
		}
	}
}

func (c ObsClient) Walk0(logger *pterm.Logger, bucket, basePath string, dirOnly bool) iter.Seq[ObsPath] {
	i := obs.ListObjectsInput{}
	i.Bucket = bucket
	i.MaxKeys = 1000
	i.EncodingType = "url"
	i.Prefix = basePath
	return c.iterPaths(logger, i, -1, dirOnly)
}

func (c ObsClient) RenameObject(logger *pterm.Logger, bucket, key, after string) {
	fullKey := fmt.Sprintf("obs://%s/%s", bucket, key)
	argsOk := logger.Args("before", fullKey, "after", fmt.Sprintf("obs://%s/%s", bucket, after))
	if strings.HasSuffix(key, "/") {
		_, err := c.RenameFolder(&obs.RenameFolderInput{
			Bucket:       bucket,
			Key:          key,
			NewObjectKey: after,
		})
		if err != nil {
			logger.Warn("unable to rename directory.", logger.Args("dir", fullKey, "error", err))
		} else {
			logger.Debug("rename directory success.", argsOk)
		}
	} else {
		_, err := c.RenameFile(&obs.RenameFileInput{
			Bucket:       bucket,
			Key:          key,
			NewObjectKey: after,
		})
		if err != nil {
			logger.Warn("unable to rename file.", logger.Args("file", fullKey, "error", err))
		} else {
			logger.Debug("rename file success.", argsOk)
		}
	}
}

func (c ObsClient) SetStorageClass(logger *pterm.Logger, bucket, key string, class obs.StorageClassType) {
	fullKey := fmt.Sprintf("obs://%s/%s", bucket, key)
	_, err := c.SetObjectMetadata(&obs.SetObjectMetadataInput{
		Bucket:            bucket,
		Key:               key,
		MetadataDirective: obs.ReplaceNew,
		StorageClass:      class,
	})
	if err != nil {
		logger.Warn("unable to set storage class for object.", logger.Args("path", fullKey, "error", err))
	} else {
		logger.Debug("set storage class for object success.", logger.Args("path", fullKey, "class", class))
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
