package clients

import (
	"fmt"
	"os"

	req "github.com/imroc/req/v3"
	"github.com/jcmturner/gokrb5/v8/client"
	"github.com/jcmturner/gokrb5/v8/config"
	"github.com/jcmturner/gokrb5/v8/credentials"
	"github.com/unhealme/lakehouse-admin-tools/internal"
	"github.com/unhealme/lakehouse-admin-tools/pkg/utils"
)

func NewKerberosClient() (*client.Client, error) {
	conf, err := config.Load(internal.GetEnv("KRB5_CONFIG", "/etc/krb5.conf"))
	if err != nil {
		return nil, err
	}
	cc, err := credentials.LoadCCache(internal.GetEnv("KRB5CCNAME", fmt.Sprintf("/tmp/krb5cc_%d", os.Getuid())))
	if err != nil {
		return nil, err
	}
	kc, err := client.NewFromCCache(cc, conf)
	if err != nil {
		return nil, err
	}
	return kc, nil
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
