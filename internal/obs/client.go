package obs

import (
	"fmt"
	"iter"
	"strings"

	"github.com/huaweicloud/huaweicloud-sdk-go-obs/obs"
	"github.com/unhealme/lakehouse-admin-tools/internal"
)

type ObsClient struct {
	*obs.ObsClient
	logger *internal.Logger
}

func (c ObsClient) Walk(bucket, basePath string, maxDepth int, dirOnly bool) iter.Seq[ObsPath] {
	fetchPaths := func(input obs.ListObjectsInput, key string, depth int) iter.Seq[ObsPath] {
		return func(yield func(ObsPath) bool) {
			p := 1
			for {
				c.logger.Info("listing obs paths.", c.logger.Args("depth", depth, "path", key, "page", p))
				r, err := c.ListObjects(&input)
				if err != nil {
					c.logger.Warn("unable to list obs paths.", c.logger.Args("depth", depth, "path", key, "error", err))
					break
				}
				c.logger.Info("obs paths fetched.", c.logger.Args("depth", depth, "path", key, "page", p, "paths", len(r.Contents)+len(r.CommonPrefixes)))
				for _, v := range r.CommonPrefixes {
					if !yield(NewPath(depth, bucket, v, nil)) {
						return
					}
				}
				if !dirOnly {
					for _, v := range r.Contents {
						if !yield(NewPath(depth, bucket, v.Key, &v)) {
							return
						}
					}
				}
				if r.IsTruncated {
					input.Marker = r.NextMarker
					p++
				} else {
					break
				}
			}
		}
	}
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
				for op := range fetchPaths(i, fmt.Sprintf("obs://%s/%s", i.Bucket, i.Prefix), depth) {
					if !yield(op) {
						return
					}
					if op.IsDir() {
						nextDepth = append(nextDepth, op.Name)
					}
				}
			}
			if len(nextDepth) < 1 {
				break
			}
		}
	}
}

func (c ObsClient) RenameObject(bucket, key, after string) {
	if strings.HasSuffix(key, "/") {
		_, err := c.RenameFolder(&obs.RenameFolderInput{Bucket: bucket, Key: key, NewObjectKey: after})
		if err != nil {
			c.logger.Warn("unable to rename directory.", c.logger.Args("dir", key, "error", err))
		} else {
			c.logger.Debug("rename directory success.", c.logger.Args("before", key, "after", after))
		}
	} else {
		_, err := c.RenameFile(&obs.RenameFileInput{Bucket: bucket, Key: key, NewObjectKey: after})
		if err != nil {
			c.logger.Warn("unable to rename file.", c.logger.Args("file", key, "error", err))
		} else {
			c.logger.Debug("rename file success.", c.logger.Args("before", key, "after", after))
		}
	}
}

func (c ObsClient) SetStorageClass(bucket, key string, class obs.StorageClassType) {
	_, err := c.SetObjectMetadata(&obs.SetObjectMetadataInput{Bucket: bucket, Key: key, MetadataDirective: obs.ReplaceNew, StorageClass: class})
	if err != nil {
		c.logger.Warn("unable to set storage class for object.", c.logger.Args("bucket", bucket, "key", key, "error", err))
	} else {
		c.logger.Debug("set storage class for object success.", c.logger.Args("bucket", bucket, "key", key, "class", class))
	}
}

func NewClient(logger *internal.Logger, endpoint string, ak, sk string, token string) (*ObsClient, error) {
	base, err := obs.New(ak, sk, endpoint,
		obs.WithSecurityToken(token),
		obs.WithSecurityProviders(obs.NewEcsSecurityProvider(1)),
	)
	if err != nil {
		return nil, err
	}
	return &ObsClient{base, logger}, nil
}
