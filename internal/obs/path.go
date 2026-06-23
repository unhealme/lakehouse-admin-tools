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
	Depth   int
	Bucket  string
	Key     string
	Name    string
	Content *obs.Content
}

func (p ObsPath) IsDir() bool {
	return strings.HasSuffix(p.Key, "/")
}

func PathFromURI(uri string) (*ObsPath, error) {
	if m := OBSURIPattern.FindStringSubmatch(uri); m != nil {
		op := NewPath(0, m[OBSURIPattern.SubexpIndex("bucket")], m[OBSURIPattern.SubexpIndex("key")], nil)
		return &op, nil
	}
	return nil, fmt.Errorf("unable to get bucket and/or key from uri: %q", uri)
}

func NewPath(depth int, bucket, key string, content *obs.Content) ObsPath {
	return ObsPath{depth, bucket, key, path.Base(key), content}
}
