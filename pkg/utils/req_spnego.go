package utils

import (
	"net/http"

	req "github.com/imroc/req/v3"
	"github.com/jcmturner/gokrb5/v8/client"
	"github.com/jcmturner/gokrb5/v8/spnego"
)

func SPNEGORedirect(req *http.Request, via []*http.Request) error {
	req.Header.Del(spnego.HTTPHeaderAuthRequest)
	return nil
}

func SPNEGORetryCond(resp *http.Response, err error) bool {
	return resp != nil && resp.StatusCode == http.StatusUnauthorized &&
		resp.Header.Get(spnego.HTTPHeaderAuthResponse) == spnego.HTTPHeaderAuthResponseValueKey
}

func SPNEGORoundTripper(kc *client.Client, spn string) req.HttpRoundTripWrapperFunc {
	return func(rt http.RoundTripper) req.HttpRoundTripFunc {
		return func(req *http.Request) (resp *http.Response, err error) {
			resp, err = rt.RoundTrip(req)
			if SPNEGORetryCond(resp, err) {
				if err = spnego.SetSPNEGOHeader(kc, resp.Request, spn); err != nil {
					return nil, err
				}
				return SPNEGORoundTripper(kc, spn)(rt)(resp.Request)
			}
			return
		}
	}
}
