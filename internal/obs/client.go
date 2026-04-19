package obs

import (
	"fmt"
	"strings"

	"github.com/huaweicloud/huaweicloud-sdk-go-obs/obs"
	"github.com/unhealme/lakehouse-admin-tools/internal"
)

var logger = internal.DefaultLogger()

type ObsClient struct{ *obs.ObsClient }

func (c *ObsClient) IterPaths(bucket, basePath string, dirOnly bool) <-chan string {
	i := &obs.ListObjectsInput{}
	i.Bucket = bucket
	i.MaxKeys = 1000
	i.EncodingType = "url"
	i.Delimiter = "/"
	i.Prefix = basePath

	path := make(chan string)
	key := fmt.Sprintf("obs://%s/%s", i.Bucket, i.Prefix)
	go func() {
		p := 1
		for {
			logger.Info("listing obs paths.", logger.Args("path", key, "page", p))
			r, err := c.ListObjects(i)
			if err != nil {
				logger.Warn("unable to list obs paths.", logger.Args("path", key, "error", err))
				break
			}
			logger.Info("obs paths fetched.", logger.Args("path", key, "page", p, "paths", len(r.Contents)+len(r.CommonPrefixes)))
			for _, v := range r.CommonPrefixes {
				path <- v
			}
			if !dirOnly {
				for _, v := range r.Contents {
					path <- v.Key
				}
			}
			if r.IsTruncated {
				i.Marker = r.NextMarker
				p++
			} else {
				break
			}
		}
		close(path)
	}()
	return path
}

func (c *ObsClient) RenameObject(bucket, key, after string) {
	if strings.HasSuffix(key, "/") {
		_, err := c.RenameFolder(&obs.RenameFolderInput{Bucket: bucket, Key: key, NewObjectKey: after})
		if err != nil {
			logger.Warn("unable to rename directory.", logger.Args("dir", key, "error", err))
		} else {
			logger.Debug("rename directory success.", logger.Args("before", key, "after", after))
		}
	} else {
		_, err := c.RenameFile(&obs.RenameFileInput{Bucket: bucket, Key: key, NewObjectKey: after})
		if err != nil {
			logger.Warn("unable to rename file.", logger.Args("file", key, "error", err))
		} else {
			logger.Debug("rename file success.", logger.Args("before", key, "after", after))
		}
	}
}

func New(endpoint string, ak, sk string, token *string) (*ObsClient, error) {
	var (
		base *obs.ObsClient
		err  error
	)
	if token != nil {
		base, err = obs.New(ak, sk, endpoint, obs.WithSecurityToken(*token))
	} else {
		base, err = obs.New(ak, sk, endpoint)
	}
	if err != nil {
		return nil, err
	}
	return &ObsClient{base}, nil
}
