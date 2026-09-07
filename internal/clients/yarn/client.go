package yarn

import (
	"bytes"
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	json "github.com/goccy/go-json"
	"github.com/jcmturner/gokrb5/v8/client"
	"github.com/jcmturner/gokrb5/v8/config"
	"github.com/jcmturner/gokrb5/v8/credentials"
	"github.com/jcmturner/gokrb5/v8/spnego"
	"github.com/pterm/pterm"
	"github.com/unhealme/lakehouse-admin-tools/internal"
)

type YarnRMClient struct {
	Http   *spnego.Client
	RmUrls []*url.URL

	krbClient *client.Client
}

func (c *YarnRMClient) Close() {
	c.Http.CloseIdleConnections()
	c.Http = nil
}

func (c *YarnRMClient) Applications(logger *pterm.Logger, states []ApplicationState, user, queue string, limit int) (*Applications, error) {
	req, err := http.NewRequest(http.MethodGet, "/ws/v1/cluster/apps", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")

	query := req.URL.Query()
	if states != nil {
		var statesString []string
		for _, state := range states {
			statesString = append(statesString, string(state))
		}
		query.Add("states", strings.Join(statesString, ","))
	}
	if user != "" {
		query.Add("user", user)
	}
	if queue != "" {
		query.Add("queue", queue)
	}
	if limit > 0 {
		query.Add("limit", strconv.Itoa(limit))
	}
	req.URL.RawQuery = query.Encode()

	logger.Debug("fetching yarn applications.")
	resp, err := c.failoverDo(logger, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var apps Applications
	if err := json.NewDecoder(resp.Body).Decode(&apps); err != nil {
		return nil, err
	}
	return &apps, nil
}

func (c *YarnRMClient) KillApplication(logger *pterm.Logger, app Application) error {
	req, err := http.NewRequest(
		http.MethodPut,
		fmt.Sprintf("/ws/v1/cluster/apps/%s/state", app.Id),
		bytes.NewBuffer([]byte(`{"state":"KILLED"}`)),
	)
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/json")

	resp, err := c.failoverDo(logger, req)
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}

func (c *YarnRMClient) failoverDo(logger *pterm.Logger, req *http.Request) (resp *http.Response, err error) {
	rmUrl := c.RmUrls[0]
	for range len(c.RmUrls) {
		newReq := *req
		newReq.URL.Scheme, newReq.URL.Host = rmUrl.Scheme, rmUrl.Host
		logger.Debug("trying url.", logger.Args("url", req.URL.String()))
		spnego.SetSPNEGOHeader(c.krbClient, &newReq, "")
		if resp, err = c.Http.Do(&newReq); err != nil {
			return nil, err
		}
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			logger.Debug("current url order:", logger.Args("url", c.getUrlOrder()))
			return
		}
		err = internal.HttpNotOkFromResponse(resp)
		resp.Body.Close()

		c.RmUrls = append(c.RmUrls[1:], rmUrl)
		rmUrl = c.RmUrls[0]
	}
	return nil, fmt.Errorf("No YARN Resource Manager is available. last error is: %s", err)
}

func (c YarnRMClient) getUrlOrder() []string {
	urls := make([]string, len(c.RmUrls))
	for _, url := range c.RmUrls {
		urls = append(urls, url.String())
	}
	return urls
}

func NewClient(rmAddresses []string) (*YarnRMClient, error) {
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

	krbConfig, err := config.Load(internal.GetEnv("KRB5_CONFIG", "/etc/krb5.conf"))
	if err != nil {
		return nil, err
	}
	krbCache, err := credentials.LoadCCache(internal.GetEnv("KRB5CCNAME", fmt.Sprintf("/tmp/krb5cc_%d", os.Getuid())))
	if err != nil {
		return nil, err
	}
	krbClient, err := client.NewFromCCache(krbCache, krbConfig)
	if err != nil {
		return nil, err
	}
	tr := http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
	httpClient := &http.Client{Transport: &tr}

	return &YarnRMClient{spnego.NewClient(krbClient, httpClient, ""), rmUrls, krbClient}, nil
}
