package hetu

import (
	"crypto/tls"
	"fmt"
	"iter"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strconv"
	"time"

	"github.com/goccy/go-json"
	"github.com/pterm/pterm"
	"github.com/unhealme/lakehouse-admin-tools/internal"
)

type HetuClient struct {
	Http    *http.Client
	HetuUrl *url.URL
	HwToken string
}

func (c HetuClient) Close() {
	c.Http.CloseIdleConnections()
}

func (c *HetuClient) GetToken() error {
	tokenUrl := c.HetuUrl.JoinPath("/v1/hsconsole/session/token")
	tokenUrl.RawQuery = fmt.Sprintf("_=%d", time.Now().UnixMilli())
	resp, err := c.Http.Get(tokenUrl.String())
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return internal.HttpNotOkFromResponse(resp)
	}
	var token HetuToken
	if err := json.NewDecoder(resp.Body).Decode(&token); err != nil {
		return err
	}
	c.HwToken = token.Token
	return nil
}

func (c HetuClient) IterTenantInfo(logger *pterm.Logger) iter.Seq[*Tenant] {
	return func(yield func(*Tenant) bool) {
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

func (c HetuClient) TenantInfo(page int) (*TenantInfo, error) {
	reqUrl := c.HetuUrl.JoinPath("/v1/hsconsole/clusters/tenant_info")
	reqQuery := reqUrl.Query()
	reqQuery.Add("size", "100")
	reqQuery.Add("page", strconv.FormatInt(int64(page), 10))
	reqQuery.Add("_", strconv.FormatInt(time.Now().UnixMilli(), 10))
	reqUrl.RawQuery = reqQuery.Encode()
	req, err := http.NewRequest(http.MethodGet, reqUrl.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("X-HW-FI-Auth-Token", c.HwToken)
	resp, err := c.Http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var tenantInfo TenantInfo
	if err := json.NewDecoder(resp.Body).Decode(&tenantInfo); err != nil {
		return nil, err
	}
	return &tenantInfo, nil
}

func NewClient(hetuAuth *HetuAuth) *HetuClient {
	tr := http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
	cookie, _ := cookiejar.New(nil)
	cookie.SetCookies(hetuAuth.Url, []*http.Cookie{hetuAuth.SessionId})
	httpClient := &http.Client{Transport: &tr, Jar: cookie}
	return &HetuClient{Http: httpClient, HetuUrl: hetuAuth.Url}
}
