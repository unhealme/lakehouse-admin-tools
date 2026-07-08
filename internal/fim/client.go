package fim

import (
	"bytes"
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strconv"
	"time"

	"github.com/goccy/go-json"
	"github.com/pterm/pterm"
	"github.com/unhealme/lakehouse-admin-tools/internal"
	"github.com/unhealme/lakehouse-admin-tools/internal/hetu"
)

type FimClient struct {
	Http    *http.Client
	FimUrl  *url.URL
	HwToken string
}

func (c FimClient) Close() {
	c.Http.CloseIdleConnections()
}

func (c FimClient) Clusters() (Clusters, error) {
	getUrl := c.FimUrl.JoinPath("/mrsmanager/api/v2/clusters")
	getUrl.RawQuery = fmt.Sprintf("_=%d", time.Now().UnixMilli())

	req, err := http.NewRequest(http.MethodGet, getUrl.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("X-HW-FI-Auth-Token", c.HwToken)
	req.AddCookie(&http.Cookie{Name: "FI_Auth_Token", Value: c.HwToken})

	resp, err := c.Http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return nil, internal.HttpNotOkFromResponse(resp)
	}

	clusters := make(Clusters, 0)
	if err := json.NewDecoder(resp.Body).Decode(&clusters); err != nil {
		return nil, err
	}
	return clusters, nil
}

func (c *FimClient) Login(loginUser, token string) error {
	body := internal.BuildUrlEncodedPayload(map[string]string{
		"eip":       c.FimUrl.Hostname(),
		"userToken": token,
		"loginUser": loginUser,
		"lan":       "en-us",
		"timestamp": strconv.FormatInt(time.Now().UnixMilli(), 10),
	})
	resp, err := c.Http.Post(
		c.FimUrl.JoinPath("/gateway/iamcert/api/v1/mrsmanager/fi-login").String(),
		"application/x-www-form-urlencoded",
		bytes.NewBufferString(body),
	)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return internal.HttpNotOkFromResponse(resp)
	}

	resp, err = c.Http.Post(
		c.FimUrl.JoinPath("/mrsmanager/api/v2/session/login_check").String(),
		"application/json; charset=UTF-8",
		nil,
	)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return internal.HttpNotOkFromResponse(resp)
	}
	var hwToken AuthToken
	if err := json.NewDecoder(resp.Body).Decode(&hwToken); err != nil {
		return err
	}
	c.HwToken = hwToken.Token
	return nil
}

func (c FimClient) getHetuEngineLinks(clusterId int) ([]string, error) {
	getUrl := c.FimUrl.JoinPath(fmt.Sprintf("/mrsmanager/api/v2/clusters/%d/services/HetuEngine/summary", clusterId))
	getUrl.RawQuery = fmt.Sprintf("_=%d", time.Now().UnixMilli())

	req, err := http.NewRequest(http.MethodGet, getUrl.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("X-HW-FI-Auth-Token", c.HwToken)
	req.AddCookie(&http.Cookie{Name: "FI_Auth_Token", Value: c.HwToken})

	resp, err := c.Http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return nil, internal.HttpNotOkFromResponse(resp)
	}

	var summary ServiceSummary
	if err := json.NewDecoder(resp.Body).Decode(&summary); err != nil {
		return nil, err
	}
	for _, property := range summary.Properties {
		if property.Key == "HSConsole WebUI" && property.Type == "LINK" {
			return property.LinkValues()
		}
	}
	return nil, errors.New("No Hetu links found.")
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
		hetuUrl := c.FimUrl.JoinPath(link)
		resp, err := c.Http.Get(hetuUrl.String())
		if err != nil {
			hetuError = err
			continue
		}
		if resp.StatusCode >= 400 {
			hetuError = internal.HttpNotOkFromResponse(resp)
			resp.Body.Close()
			continue
		}
		resp.Body.Close()

		hetuAuth.Url = hetuUrl
		for _, cookie := range c.Http.Jar.Cookies(hetuUrl) {
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

func NewClient(fimAddress string) (*FimClient, error) {
	url, err := url.Parse(fimAddress)
	if err != nil {
		return nil, err
	}
	tr := http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
	cookie, _ := cookiejar.New(nil)
	httpClient := &http.Client{Transport: &tr, Jar: cookie}
	return &FimClient{Http: httpClient, FimUrl: url}, err
}
