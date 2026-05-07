package obs

import (
	"fmt"
	"iter"
	"strings"

	"github.com/huaweicloud/huaweicloud-sdk-go-obs/obs"
	"github.com/unhealme/lakehouse-admin-tools/internal"
)

type ObsClient struct{ *obs.ObsClient }

func (c ObsClient) iterPaths(logger *internal.Logger, input obs.ListObjectsInput, depth int, dirOnly bool) iter.Seq[ObsPath] {
	return func(yield func(ObsPath) bool) {
		p := 1
		for {
			logArgs := []any{"depth", depth, "path", fmt.Sprintf("obs://%s/%s", input.Bucket, input.Prefix), "page", p}
			logger.Debug("listing obs paths.", logger.Args(logArgs...))
			r, err := c.ListObjects(&input)
			if err != nil {
				logger.Warn("unable to list obs paths.", logger.Args(append(logArgs, "error", err)...))
				break
			}
			logger.Debug("obs paths fetched.", logger.Args(append(logArgs, "count", len(r.Contents)+len(r.CommonPrefixes))...))
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

func (c ObsClient) Walk(logger *internal.Logger, bucket, basePath string, maxDepth int, dirOnly bool) iter.Seq[ObsPath] {
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

func (c ObsClient) RenameObject(logger *internal.Logger, bucket, key, after string) {
	fullKey := fmt.Sprintf("obs://%s/%s", bucket, key)
	argsOk := logger.Args("before", fullKey, "after", fmt.Sprintf("obs://%s/%s", bucket, after))
	if strings.HasSuffix(key, "/") {
		_, err := c.RenameFolder(&obs.RenameFolderInput{Bucket: bucket, Key: key, NewObjectKey: after})
		if err != nil {
			logger.Warn("unable to rename directory.", logger.Args("dir", fullKey, "error", err))
		} else {
			logger.Debug("rename directory success.", argsOk)
		}
	} else {
		_, err := c.RenameFile(&obs.RenameFileInput{Bucket: bucket, Key: key, NewObjectKey: after})
		if err != nil {
			logger.Warn("unable to rename file.", logger.Args("file", fullKey, "error", err))
		} else {
			logger.Debug("rename file success.", argsOk)
		}
	}
}

func (c ObsClient) SetStorageClass(logger *internal.Logger, bucket, key string, class obs.StorageClassType) {
	fullKey := fmt.Sprintf("obs://%s/%s", bucket, key)
	_, err := c.SetObjectMetadata(&obs.SetObjectMetadataInput{Bucket: bucket, Key: key, MetadataDirective: obs.ReplaceNew, StorageClass: class})
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
	)
	if err != nil {
		return nil, err
	}
	return &ObsClient{base}, nil
}
