package mrs

import (
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/auth"
	mrs "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/mrs/v2"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/mrs/v2/region"
)

type MrsClient struct{ *mrs.MrsClient }

func (c MrsClient) GetClusterManagerToken(clusterId string) (*GetClusterManagerTokenResponse, error) {
	request := &GetClusterManagerTokenRequest{clusterId}
	requestDef := GenReqDefForListClusterManagerAuthState()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*GetClusterManagerTokenResponse), nil
	}
}

func NewClient(ak, sk, token, regionId string) (*MrsClient, error) {
	credsBuilder := auth.NewBasicCredentialsBuilder().WithAk(ak).WithSk(sk)
	if token != "" {
		credsBuilder = credsBuilder.WithSecurityToken(token)
	}
	creds, err := credsBuilder.SafeBuild()
	if err != nil {
		return nil, err
	}
	region, err := region.SafeValueOf(regionId)
	if err != nil {
		return nil, err
	}
	base, err := mrs.MrsClientBuilder().WithCredential(creds).WithRegion(region).SafeBuild()
	if err != nil {
		return nil, err
	}
	return &MrsClient{mrs.NewMrsClient(base)}, nil
}
