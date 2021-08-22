package http

import (
	"encoding/json"
	"fmt"
	stdhttp "net/http"
	"strings"
)

type Response struct {
	Headers stdhttp.Header
	Status  int
	body    responseReader
}

func (r *Response) Read(p []byte) (n int, err error) {
	return r.body.Read(p)
}

func (r *Response) Reset() {
	r.body.Reset()
}

func (r *Response) Close() error {
	return r.body.Close()
}

// ResponseOption is the option type for responses
type ResponseOption interface {
	ProcessResponse(*Response) error
}

func (j JSONOption) ProcessResponse(resp *Response) error {
	ct := resp.Headers.Get("Content-Type")
	if ct != "" && !strings.HasPrefix(ct, "application/json") {
		return fmt.Errorf("invalid Content-Type header, expected 'application/json', got %s", ct)
	}
	return json.NewDecoder(resp.body).Decode(j.v)
}

func (resp *Response) applyOptions(options ...ResponseOption) error {
	for _, opt := range options {
		if err := opt.ProcessResponse(resp); err != nil {
			return err
		}
	}
	return nil
}
