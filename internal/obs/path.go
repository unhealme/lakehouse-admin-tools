package obs

import (
	"path"
	"strings"

	"github.com/huaweicloud/huaweicloud-sdk-go-obs/obs"
)

type ObsPath struct {
	Depth   int
	Bucket  string
	Key     string
	Name    string
	Content *obs.Content
}

func (p ObsPath) IsDir() bool {
	return strings.HasSuffix(p.Key, "/")
}

func NewPath(depth int, bucket, key string, content *obs.Content) ObsPath {
	return ObsPath{depth, bucket, key, path.Base(key), content}
}
