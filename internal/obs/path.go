package obs

import (
	"fmt"
	"path"
	"regexp"
	"strings"

	"github.com/huaweicloud/huaweicloud-sdk-go-obs/obs"
)

var OBSURIPattern = regexp.MustCompile(`^obs://(?P<bucket>[^/]+)/(?P<key>.+)$`)

type ObsPath struct {
	Bucket string `json:"bucket"`
	Key    string `json:"key"`
}

func (p ObsPath) IsDir() bool {
	return strings.HasSuffix(p.Key, "/")
}

func (p ObsPath) Name() string {
	return path.Base(p.Key)
}

func (p ObsPath) URI() string {
	return fmt.Sprintf("obs://%s/%s", p.Bucket, p.Key)
}

func (p ObsPath) WithKey(key string) ObsPath {
	return NewObsPath(p.Bucket, key)
}

func NewObsPath(bucket, key string) ObsPath {
	return ObsPath{bucket, key}
}

func PathFromURI(uri string) (*ObsPath, error) {
	if m := OBSURIPattern.FindStringSubmatch(uri); m != nil {
		op := NewObsPath(m[OBSURIPattern.SubexpIndex("bucket")], m[OBSURIPattern.SubexpIndex("key")])
		return &op, nil
	}
	return nil, fmt.Errorf("unable to get bucket and/or key from uri: %q", uri)
}

type ObsPathContent struct {
	ObsPath
	Depth   int
	Content *obs.Content
}

func NewObsPathContent(depth int, bucket, key string, content *obs.Content) ObsPathContent {
	v := ObsPathContent{Depth: depth, Content: content}
	v.Bucket = bucket
	v.Key = key
	return v
}

type ObsPathAnalyzed struct {
	ObsPath
	Exists    bool  `json:"exists"`
	DirCount  int   `json:"dir_count"`
	FileCount int   `json:"file_count"`
	Size      int64 `json:"size"`
}
