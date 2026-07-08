package internal

import (
	"fmt"
	"io"
	"net/http"
)

type HttpNotOk struct {
	Status int
	Header http.Header
	Err    error
	Body   []byte
}

func (e HttpNotOk) Error() string {
	return fmt.Sprintf("HTTP%d(headers=%s, error=%s, body=%s)", e.Status, e.Header, e.Err, e.Body)
}

func HttpNotOkFromResponse(response *http.Response) error {
	body, err := io.ReadAll(response.Body)
	return HttpNotOk{
		Status: response.StatusCode,
		Header: response.Header,
		Err:    err,
		Body:   body,
	}
}
