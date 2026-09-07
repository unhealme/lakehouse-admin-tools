package dataarts

var (
	EnvTypeProd int32 = 1
	EnvTypeDev  int32 = 2

	SupportServiceDefault int32 = 108

	CreateConnectionsRequestBodyModeDefault    int32 = 1
	CreateConnectionsRequestBodyVisibleDefault int32 = 1
)

type DwConfig struct {
	Manually             *bool   `json:"manually,omitempty"`
	ClusterId            *string `json:"clusterId,omitempty"`
	ClusterName          *string `json:"clusterName,omitempty"`
	DbUserName           *string `json:"dbUserName,omitempty"`
	DbPassword           *string `json:"dbPassword"`
	RememberPassword     *bool   `json:"rememberPassword,omitempty"`
	KmsId                *string `json:"kmsId,omitempty"`
	KmsName              *string `json:"kmsName,omitempty"`
	HsBrokerIps          *string `json:"hsBrokerIps,omitempty"`
	HsBrokerPort         *string `json:"hsBrokerPort,omitempty"`
	Properties           *string `json:"properties,omitempty"`
	CdmPropertyEnable    *bool   `json:"cdmPropertyEnable,omitempty"`
	EnableRealtimeSync   *bool   `json:"metadata.enableRealtimeSync,omitempty"`
	EnableAutoCollection *bool   `json:"metadata.enableAutoCollection,omitempty"`
}
