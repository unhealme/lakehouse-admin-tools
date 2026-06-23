package yarn

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/goccy/go-json"
	"github.com/jcmturner/gokrb5/v8/client"
	"github.com/jcmturner/gokrb5/v8/config"
	"github.com/jcmturner/gokrb5/v8/credentials"
	"github.com/jcmturner/gokrb5/v8/spnego"
	"github.com/unhealme/lakehouse-admin-tools/internal"
)

type YarnRMClient struct {
	*spnego.Client
	RMAddress string
}

func (c YarnRMClient) Applications(logger *internal.Logger, states []ApplicationState, user, queue string, limit int) (*Applications, error) {
	req, err := http.NewRequest(http.MethodGet, c.RMAddress+"/ws/v1/cluster/apps", nil)
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

	logger.Debug("fetching yarn applications.", logger.Args("url", req.URL.String()))
	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, internal.HttpNotOk{Status: resp.StatusCode, Header: resp.Header, Err: err, Body: nil}
		}
		return nil, internal.HttpNotOk{Status: resp.StatusCode, Header: resp.Header, Err: err, Body: body}
	}

	var apps Applications
	if err := json.NewDecoder(resp.Body).Decode(&apps); err != nil {
		return nil, err
	}
	return &apps, nil
}

func (c YarnRMClient) KillApplication(logger *internal.Logger, app Application) bool {
	req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("%s/ws/v1/cluster/apps/%s/state", c.RMAddress, app.Id), bytes.NewBuffer([]byte(`{"state":"KILLED"}`)))
	if err != nil {
		logger.Error("unable to kill yarn application.", logger.Args("id", app.Id, "error", err))
		return false
	}
	req.Header.Add("Content-Type", "application/json")

	resp, err := c.Do(req)
	if err != nil {
		logger.Error("unable to kill yarn application.", logger.Args("id", app.Id, "error", err))
		return false
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			logger.Error("unable to kill yarn application.", logger.Args("id", app.Id, "header", resp.Header, "error", err))
			return false
		}
		logger.Error("unable to kill yarn application.", logger.Args("id", app.Id, "header", resp.Header, "error", err, "body", string(body)))
		return false
	}
	return true
}

func NewClient(rmAddress string) (*YarnRMClient, error) {
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
	spnegoClient := spnego.NewClient(krbClient, httpClient, "")
	return &YarnRMClient{spnegoClient, strings.TrimRight(rmAddress, "/")}, nil
}
