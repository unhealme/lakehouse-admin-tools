package clients

import (
	"fmt"
	"os"
	"strings"

	req "github.com/imroc/req/v3"
	"github.com/jcmturner/gokrb5/v8/client"
	"github.com/jcmturner/gokrb5/v8/config"
	"github.com/jcmturner/gokrb5/v8/credentials"
	"github.com/unhealme/lakehouse-admin-tools/internal"
	"github.com/unhealme/lakehouse-admin-tools/pkg/utils"
)

func NewKerberosClient() (kc *client.Client, err error) {
	var (
		conf *config.Config
		cc   *credentials.CCache
	)
	if conf, err = config.Load(internal.GetEnv("KRB5_CONFIG", "/etc/krb5.conf")); err == nil {
		if cc, err = credentials.LoadCCache(
			strings.TrimPrefix(internal.GetEnv(
				"KRB5CCNAME",
				fmt.Sprintf("/tmp/krb5cc_%d", os.Getuid()),
			), "FILE:"),
		); err == nil {
			kc, err = client.NewFromCCache(cc, conf)
		}
	}
	return
}

func NewKerberosHttpClient(spn string) (*req.Client, error) {
	kc, err := NewKerberosClient()
	if err != nil {
		return nil, err
	}
	c := NewHttpClient().SetRedirectPolicy(utils.SPNEGORedirect)
	c.GetTransport().WrapRoundTripFunc(utils.SPNEGORoundTripper(kc, spn))
	return c, nil
}
