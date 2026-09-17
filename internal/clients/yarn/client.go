package yarn

import (
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	req "github.com/imroc/req/v3"
	"github.com/jcmturner/gokrb5/v8/spnego"
	"github.com/pterm/pterm"
	"github.com/unhealme/lakehouse-admin-tools/internal/clients"
)

type YarnRMClient struct {
	Http   *req.Client
	RmUrls []*url.URL
}

func (c *YarnRMClient) Close() {
	c.Http.CloseIdleConnections()
	c.Http = nil
}

func (c *YarnRMClient) Applications(logger *pterm.Logger, states []ApplicationState, user, queue string, limit int) (*Applications, error) {
	var apps Applications
	req := c.Http.R().SetHeader("Content-Type", "application/json").SetSuccessResult(&apps)
	if len(states) > 0 {
		var statesString []string
		for _, state := range states {
			statesString = append(statesString, string(state))
		}
		req.SetQueryParam("states", strings.Join(statesString, ","))
	}
	if user != "" {
		req.SetQueryParam("user", user)
	}
	if queue != "" {
		req.SetQueryParam("queue", queue)
	}
	if limit > 0 {
		req.SetQueryParam("limit", strconv.Itoa(limit))
	}

	logger.Debug("fetching yarn applications.")
	if _, err := req.Get("/ws/v1/cluster/apps"); err != nil {
		return nil, err
	}
	return &apps, nil
}

var killAppBody = []byte(`{"state":"KILLED"}`)

func (c *YarnRMClient) KillApplication(logger *pterm.Logger, app Application) error {
	_, err := c.Http.R().
		SetBody(killAppBody).
		SetHeader("Content-Type", "application/json").
		Get(fmt.Sprintf("/ws/v1/cluster/apps/%s/state", app.Id))
	return err
}

func (c *YarnRMClient) failoverRetry(logger *pterm.Logger) req.RetryHookFunc {
	return func(resp *req.Response, err error) {
		c.RmUrls = append(c.RmUrls[1:], c.RmUrls[0])
		rmUrl := c.RmUrls[0]
		c.Http.SetBaseURL(rmUrl.String())

		lastReq := resp.Request
		newUrl := lastReq.URL.Clone()
		newUrl.Scheme, newUrl.Host = rmUrl.Scheme, rmUrl.Host
		lastReq.RawURL = newUrl.String()
		lastReq.Headers.Del(spnego.HTTPHeaderAuthRequest)
		logger.Debug("swapped urls with new order: " + strings.Join(c.getUrlOrder(), ", "))
	}
}

func (c YarnRMClient) getUrlOrder() []string {
	urls := make([]string, len(c.RmUrls))
	for i, url := range c.RmUrls {
		urls[i] = url.String()
	}
	return urls
}

func NewClient(logger *pterm.Logger, rmAddresses []string) (*YarnRMClient, error) {
	if len(rmAddresses) < 1 {
		return nil, errors.New("Atleast one RM Address must be specified")
	}

	var rmUrls []*url.URL
	for _, rmAddress := range rmAddresses {
		rmUrl, err := url.Parse(rmAddress)
		if err != nil {
			return nil, err
		}
		rmUrls = append(rmUrls, rmUrl)
	}

	hc, err := clients.NewKerberosHttpClient("")
	if err != nil {
		return nil, err
	}
	c := new(YarnRMClient{hc, rmUrls})
	hc.SetBaseURL(rmUrls[0].String()).
		SetCommonRetryCount(len(rmUrls) - 1).
		SetCommonRetryHook(c.failoverRetry(logger))
	return c, nil
}
