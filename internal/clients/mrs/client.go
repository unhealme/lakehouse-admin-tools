package mrs

import (
	"net/url"
	"strconv"

	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/auth"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/config"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/impl"
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

func NewClient(ak, sk, token, regionId, proxy string) (*MrsClient, error) {
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

	httpConfig := config.DefaultHttpConfig()
	newCreds, err := creds.ProcessAuthParams(impl.NewDefaultHttpClient(httpConfig), region.Id)
	if err != nil {
		return nil, err
	}
	if proxy != "" {
		proxyUrl, err := url.Parse(proxy)
		if err != nil {
			return nil, err
		}
		port, _ := strconv.ParseInt(proxyUrl.Port(), 10, 32)
		httpProxy := config.NewProxy()
		httpProxy.Schema = proxyUrl.Scheme
		httpProxy.Host = proxyUrl.Hostname()
		httpProxy.Port = int(port)
		if proxyUrl.User != nil {
			httpProxy.Username = proxyUrl.User.Username()
			httpProxy.Password, _ = proxyUrl.User.Password()
		}
		httpConfig.WithProxy(httpProxy)
	}
	base, err := mrs.MrsClientBuilder().WithHttpConfig(httpConfig).WithCredential(newCreds).WithRegion(region).SafeBuild()
	if err != nil {
		return nil, err
	}
	return &MrsClient{mrs.NewMrsClient(base)}, nil
}
