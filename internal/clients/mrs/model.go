package mrs

import (
	"strings"

	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/utils"
)

type GetClusterManagerTokenResponse struct {
	Token          string `json:"token"`
	HttpStatusCode int    `json:"-"`
}

func (o GetClusterManagerTokenResponse) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "GetClusterManagerTokenResponse struct{}"
	}
	return strings.Join([]string{"GetClusterManagerTokenResponse", string(data)}, " ")
}

type GetClusterManagerTokenRequest struct {
	ClusterId string `json:"cluster_id"`
}

func (o GetClusterManagerTokenRequest) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "GetClusterManagerTokenRequest struct{}"
	}
	return strings.Join([]string{"GetClusterManagerTokenRequest", string(data)}, " ")
}
