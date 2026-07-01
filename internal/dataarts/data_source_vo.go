package dataarts

var (
	EnvTypeProd int32 = 1
	EnvTypeDev  int32 = 2

	SupportServiceDefault int32 = 108

	CreateConnectionsRequestBodyModeDefault    int32 = 1
	CreateConnectionsRequestBodyVisibleDefault int32 = 1
)

type DwConfig struct {
	Manually             bool   `json:"manually"`
	ClusterId            string `json:"clusterId"`
	ClusterName          string `json:"clusterName"`
	DbUserName           string `json:"dbUserName"`
	DbPassword           string `json:"dbPassword"`
	RememberPassword     bool   `json:"rememberPassword"`
	KmsId                string `json:"kmsId"`
	KmsName              string `json:"kmsName"`
	HsBrokerIps          string `json:"hsBrokerIps"`
	HsBrokerPort         string `json:"hsBrokerPort"`
	CdmPropertyEnable    bool   `json:"cdmPropertyEnable"`
	EnableRealtimeSync   bool   `json:"metadata.enableRealtimeSync"`
	EnableAutoCollection bool   `json:"metadata.enableAutoCollection"`
}
