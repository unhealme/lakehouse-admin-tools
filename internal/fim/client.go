package fim

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	json "github.com/goccy/go-json"
	req "github.com/imroc/req/v3"
	"github.com/pterm/pterm"
	"github.com/tidwall/gjson"
	"github.com/unhealme/lakehouse-admin-tools/internal"
	"github.com/unhealme/lakehouse-admin-tools/internal/hetu"
)

type FimClient struct {
	Http    *req.Client
	FimUrl  *url.URL
	HwToken string
}

func (c *FimClient) BasicLogin(user, passw string) error {
	c.Http.SetCommonBasicAuth(user, passw)
	return c.getToken()
}

func (c *FimClient) Close() {
	c.Http.ClearCookies().CloseIdleConnections()
	c.Http = nil
	c.HwToken = ""
}

func (c FimClient) Clusters() (clusters Clusters, err error) {
	_, err = c.Http.R().
		SetQueryString(fmt.Sprintf("_=%d", time.Now().UnixMilli())).
		SetSuccessResult(&clusters).
		Get("/mrsmanager/api/v2/clusters")
	return
}

func (c FimClient) GetHetuEngineAuth(logger *pterm.Logger, clusterId int) (*hetu.HetuAuth, error) {
	links, err := c.getHetuEngineLinks(clusterId)
	if err != nil {
		return nil, err
	}

	var (
		hetuAuth  hetu.HetuAuth
		hetuError error
	)
	for _, link := range links {
		logger.Debug("trying hetu link.", logger.Args("link", link))
		if _, err := c.Http.R().Get(link); err != nil {
			hetuError = err
			continue
		}

		hetuUrl := c.FimUrl.JoinPath(link)
		hetuAuth.Url = hetuUrl
		cookies, _ := c.Http.GetCookies(hetuUrl.String())
		for _, cookie := range cookies {
			if cookie.Name == "JSESSIONID" {
				hetuAuth.SessionId = cookie
				break
			}
		}
		logger.Debug("got hetu auth.", logger.Args("url", hetuUrl.String()))
		return &hetuAuth, nil
	}
	return nil, hetuError
}

func (c *FimClient) MrsLogin(loginUser, token string) (err error) {
	if _, err = c.Http.R().
		SetFormData(map[string]string{
			"eip":       c.FimUrl.Hostname(),
			"userToken": token,
			"loginUser": loginUser,
			"lan":       "en-us",
			"timestamp": strconv.FormatInt(time.Now().UnixMilli(), 10),
		}).
		Post("/gateway/iamcert/api/v1/mrsmanager/fi-login"); err != nil {
		return
	}
	if err = c.getToken(); err != nil {
		return
	}
	c.Http.SetCommonCookies(&http.Cookie{Name: "FI_Auth_Token", Value: c.HwToken})
	return
}

func (c FimClient) ResetUserPassword(user, passw string) (err error) {
	_, err = c.Http.R().
		SetBodyJsonMarshal(map[string]string{"newPassword": passw}).
		Post(fmt.Sprintf("%s://%s:28443/web/api/v2/permission/users/%s/password/reset",
			c.FimUrl.Scheme, c.FimUrl.Hostname(), user,
		))
	return
}

func (c FimClient) getHetuEngineLinks(clusterId int) ([]string, error) {
	var summary ServiceSummary
	if _, err := c.Http.R().
		SetQueryString(fmt.Sprintf("_=%d", time.Now().UnixMilli())).
		SetSuccessResult(&summary).
		Get(fmt.Sprintf("/mrsmanager/api/v2/clusters/%d/services/HetuEngine/summary", clusterId)); err != nil {
		return nil, err
	}
	for _, property := range summary.Properties {
		if property.Key == "HSConsole WebUI" && property.Type == "LINK" {
			return property.LinkValues()
		}
	}
	return nil, errors.New("No Hetu links found")
}

func (c *FimClient) getToken() error {
	resp, err := c.Http.R().
		SetContentType("application/json; charset=UTF-8").
		Post("/mrsmanager/api/v2/session/login_check")
	if err != nil {
		return err
	}
	token := gjson.GetBytes(resp.Bytes(), "token").String()
	c.Http.SetCommonHeader("X-HW-FI-Auth-Token", token)
	c.HwToken = token
	return nil
}

func NewClient(fimAddress string) (*FimClient, error) {
	url, err := url.Parse(fimAddress)
	if err != nil {
		return nil, err
	}
	httpClient := req.C().
		DisableAutoDecode().
		EnableInsecureSkipVerify().
		OnAfterResponse(internal.HttpNotOkMiddleware).
		SetBaseURL(url.String()).
		SetJsonMarshal(json.Marshal).
		SetJsonUnmarshal(json.Unmarshal)
	return &FimClient{Http: httpClient, FimUrl: url}, nil
}
