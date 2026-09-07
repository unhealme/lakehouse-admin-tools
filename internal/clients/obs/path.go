package obs

import (
	"fmt"
	"path"
	"regexp"
	"strconv"
	"strings"
	"time"

	json "github.com/goccy/go-json"
	"github.com/huaweicloud/huaweicloud-sdk-go-obs/obs"
	"github.com/unhealme/lakehouse-admin-tools/pkg/utils"
)

var OBSURIPattern = regexp.MustCompile(`^obs://(?P<bucket>[^/]+)/(?P<key>.+)$`)

type ObsPath struct {
	Bucket string `json:"bucket"`
	Key    string `json:"key"`
}

func (p ObsPath) Equal(o ObsPath) bool {
	if p.Bucket != o.Bucket {
		return false
	}
	return SameObsKey(p.Key, o.Key)
}

func (p ObsPath) IsDir() bool {
	return strings.HasSuffix(p.Key, "/")
}

func (p ObsPath) Name() string {
	return path.Base(p.Key)
}

func (p ObsPath) URI() string {
	return "obs://" + p.Bucket + "/" + p.Key
}

func (p ObsPath) WithKey(key string) ObsPath {
	return NewObsPath(p.Bucket, strings.TrimPrefix(key, "/"))
}

func NewObsPath(bucket, key string) ObsPath {
	return ObsPath{strings.Trim(bucket, "/"), strings.TrimPrefix(key, "/")}
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
	v.Key = strings.TrimPrefix(key, "/")
	return v
}

var ObsPathAnalyzedHeader = []string{
	"ObsPath",
	"RawSize",
	"Size",
	"DirCount",
	"FileCount",
	"LastModified",
}

type ObsPathAnalyzed struct {
	ObsPath
	Exists       bool  `json:"exists"`
	DirCount     int64 `json:"dir_count"`
	FileCount    int64 `json:"file_count"`
	Size         int64 `json:"size"`
	LastModified int64 `json:"last_modified"` // unix_milli
}

func (p ObsPathAnalyzed) SerCsv() []string {
	return []string{
		p.URI(),
		strconv.FormatInt(p.Size, 10),
		utils.FormatSize(p.Size),
		strconv.FormatInt(p.DirCount, 10),
		strconv.FormatInt(p.FileCount, 10),
		time.Unix(0, p.LastModified*int64(time.Millisecond)).Format("2006-01-02 15:04:05.000"),
	}
}

func (p ObsPathAnalyzed) SerJson() (v []byte) {
	v, _ = json.Marshal(p)
	return
}

func SameObsKey(x, y string) bool {
	return strings.TrimSuffix(x, "/") == strings.TrimSuffix(y, "/")
}
