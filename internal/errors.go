package internal

import "net/http"

type HttpNotOk struct {
	Status int
	Header http.Header
	Err    error
	Body   []byte
}

func (e HttpNotOk) Error() string {
	return "http not ok."
}
