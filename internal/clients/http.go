package clients

import (
	req "github.com/imroc/req/v3"
	"github.com/unhealme/lakehouse-admin-tools/internal"
)

func NewHttpClient() *req.Client {
	return req.C().
		DisableAutoDecode().
		EnableInsecureSkipVerify().
		OnAfterResponse(internal.HttpNotOkMiddleware)
}
