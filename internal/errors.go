package internal

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

	req "github.com/imroc/req/v3"
)

type HttpNotOk struct {
	Status  int
	Headers string
	Err     error
	Body    []byte
}

func (e HttpNotOk) Error() string {
	return fmt.Sprintf("HTTP%d(headers=%s, error=%s, body=%s)", e.Status, e.Headers, e.Err, e.Body)
}

func HttpNotOkFromResponse(response *http.Response) error {
	body, err := io.ReadAll(response.Body)
	headers := new(bytes.Buffer)
	response.Header.Write(headers)
	return HttpNotOk{
		Status:  response.StatusCode,
		Headers: headers.String(),
		Err:     err,
		Body:    body,
	}
}

func HttpNotOkMiddleware(client *req.Client, resp *req.Response) error {
	if resp.Err == nil && resp.StatusCode >= 400 {
		body, err := resp.ToBytes()
		return HttpNotOk{
			Status:  resp.StatusCode,
			Headers: resp.HeaderToString(),
			Err:     err,
			Body:    body,
		}
	}
	return nil
}
