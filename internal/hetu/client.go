package hetu

import (
	"fmt"
	"iter"
	"net/url"
	"strconv"
	"time"

	json "github.com/goccy/go-json"
	req "github.com/imroc/req/v3"
	"github.com/pterm/pterm"
	"github.com/tidwall/gjson"
	"github.com/unhealme/lakehouse-admin-tools/internal"
)

type HetuClient struct {
	Http    *req.Client
	HetuUrl *url.URL
	HwToken string
}

func (c *HetuClient) Close() {
	c.Http.ClearCookies().CloseIdleConnections()
	c.Http = nil
	c.HwToken = ""
}

func (c HetuClient) Clusters(page int) (*ClustersResponse[ClusterContent], error) {
	var clusters ClustersResponse[ClusterContent]
	if _, err := c.Http.R().
		SetQueryParam("size", "100").
		SetQueryParam("page", strconv.FormatInt(int64(page), 10)).
		SetQueryString(queryTs()).
		SetSuccessResult(&clusters).
		Get("/v1/hsconsole/clusters"); err != nil {
		return nil, err
	}
	return &clusters, nil
}

func (c HetuClient) ClustersRaw() (*ClustersResponse[ClusterContentRaw], error) {
	var (
		allClusters ClustersResponse[ClusterContentRaw]

		page  = 1
		total = 0
	)
	for {
		var clusters ClustersResponse[ClusterContentRaw]
		if _, err := c.Http.R().
			SetQueryParam("size", "100").
			SetQueryParam("page", strconv.FormatInt(int64(page-1), 10)).
			SetQueryString(queryTs()).
			SetSuccessResult(&clusters).
			Get("/v1/hsconsole/clusters"); err != nil {
			continue
		}
		if page > 1 {
			for _, cluster := range clusters.Content.Clusters {
				allClusters.Content.Clusters = append(allClusters.Content.Clusters, cluster)
				total++
			}
		} else {
			allClusters = clusters
		}
		if total >= clusters.Content.Total {
			break
		}
		page++
	}
	return &allClusters, nil
}

func (c *HetuClient) GetToken() error {
	resp, err := c.Http.R().
		SetQueryString(queryTs()).
		Get("/v1/hsconsole/session/token")
	if err != nil {
		return err
	}
	token := gjson.GetBytes(resp.Bytes(), "token").String()
	c.Http.SetCommonHeader("X-HW-FI-Auth-Token", token)
	c.HwToken = token
	return nil
}

func (c HetuClient) IterCluster(logger *pterm.Logger) iter.Seq[*Cluster] {
	return func(yield func(*Cluster) bool) {
		page := 1
		total := 0
		for {
			clusters, err := c.Clusters(page - 1)
			if err != nil {
				logger.Error("unable to get clusters.", logger.Args("page", page, "error", err))
				break
			}
			for _, cluster := range clusters.Content.Clusters {
				if !yield(&cluster) {
					return
				}
				total++
			}
			if total >= clusters.Content.Total {
				break
			}
			page++
		}
	}
}

func (c HetuClient) IterTenantInfo(logger *pterm.Logger) iter.Seq[*TenantInfo] {
	return func(yield func(*TenantInfo) bool) {
		page := 1
		total := 0
		for {
			tenantInfo, err := c.TenantInfo(page - 1)
			if err != nil {
				logger.Error("unable to get tenant info.", logger.Args("page", page, "error", err))
				break
			}
			for _, tenant := range tenantInfo.Content.Tenants {
				if !yield(&tenant) {
					return
				}
				total++
			}
			if total >= tenantInfo.Content.Total {
				break
			}
			page++
		}
	}
}

func (c HetuClient) TenantInfo(page int) (*TenantInfoResponse, error) {
	var tenantInfo TenantInfoResponse
	if _, err := c.Http.R().
		SetQueryParam("size", "100").
		SetQueryParam("page", strconv.FormatInt(int64(page), 10)).
		SetQueryString(queryTs()).
		SetSuccessResult(&tenantInfo).
		Get("/v1/hsconsole/clusters/tenant_info"); err != nil {
		return nil, err
	}
	return &tenantInfo, nil
}

func (c HetuClient) TenantConfig(tenant string) (*TenantConfigResponse, error) {
	var tenantConfig TenantConfigResponse
	if _, err := c.Http.R().
		SetQueryString(queryTs()).
		SetSuccessResult(&tenantConfig).
		Get(fmt.Sprintf("/v1/hsconsole/clusters/config/tenant/%s", tenant)); err != nil {
		return nil, err
	}
	return &tenantConfig, nil
}

func queryTs() string {
	return "_=" + strconv.FormatInt(time.Now().UnixMilli(), 64)
}

func NewClient(hetuAuth *HetuAuth) *HetuClient {
	httpClient := req.C().
		DisableAutoDecode().
		EnableInsecureSkipVerify().
		OnAfterResponse(internal.HttpNotOkMiddleware).
		SetBaseURL(hetuAuth.Url.String()).
		SetCommonCookies(hetuAuth.SessionId).
		SetJsonMarshal(json.Marshal).
		SetJsonUnmarshal(json.Unmarshal)
	return &HetuClient{Http: httpClient, HetuUrl: hetuAuth.Url}
}
