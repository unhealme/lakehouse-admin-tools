package hetu

import (
	"encoding/json"
	"net/http"
	"net/url"
)

type HetuAuth struct {
	SessionId *http.Cookie
	Url       *url.URL
}

type ClustersResponse[T ClusterContent | ClusterContentRaw] struct {
	Result  json.RawMessage `json:"result"`
	Content T               `json:"content"`
	Code    json.RawMessage `json:"code"`
	Msg     json.RawMessage `json:"msg"`
}

type ClusterContent struct {
	Clusters []Cluster `json:"clusters"`
	Total    int       `json:"total"`
}

type ClusterContentRaw struct {
	Clusters []json.RawMessage `json:"clusters"`
	Total    int               `json:"total"`
}

type Cluster struct {
	ComputerClusterId     string // hex32
	Name                  string
	State                 string
	Tenant                string
	StartTime             string // %Y-%m-%d %H:%M:%S
	FinishTime            string
	ContainerAttemptUrl   string
	WorkerAttemptUrl      string
	ComputerClusterConfig ConfigContent
	ContainerList         []Container
}

func (c Cluster) TotalMemory() (total int) {
	for _, cont := range c.ContainerList {
		total += cont.Memory
	}
	return
}

func (c Cluster) TotalVcores() (total int) {
	for _, cont := range c.ContainerList {
		total += cont.Vcores
	}
	return
}

type Container struct {
	ContainerId string
	Type        string // COORDINATOR, WORKER
	State       string
	StartTime   string // unix milli string
	Memory      int    // Megabyte
	Vcores      int

	// Unknwon Type:
	// Resource
}

type TenantInfoResponse struct {
	Content InfoContent
}

type InfoContent struct {
	Tenants []TenantInfo
	Total   int
}

type TenantInfo struct {
	Tenant       string
	TotalMemory  int // Megabyte
	TotalVcores  int
	ClusterIds   []string
	RunningCount int
	StoppedCount int
	ErrorCount   int
}

type TenantConfigResponse struct {
	Content ConfigContent
}

type ConfigContent struct {
	Name         string
	CnMemory     int // Coordinator:Megabyte
	Cn           int // Coordinator Nodes
	CnVcores     int // Coordinator Vcores
	WorkerMemory int // Megabyte
	InitWorker   int
	TargetWorker int
	CurWorker    int
	WorkerVcores int
	InstanceNum  int
	SeniorConfig SeniorConfig
}

type SeniorConfig struct {
	WorkerConfigProperties      ConfigProperties `json:"worker.config.properties"`
	CoordinatorConfigProperties ConfigProperties `json:"coordinator.config.properties"`
	WorkerJvmConfig             JvmConfig        `json:"worker.jvm.config"`
	CoordinatorJvmConfig        JvmConfig        `json:"coordinator.jvm.config"`
}

type ConfigProperties struct {
	QueryMaxMemory      string `json:"query.max-memory"`
	QueryMaxTotalMemory string `json:"query.max-total-memory"`
}

type JvmConfig struct {
	ExtraJavaOptions string
}
