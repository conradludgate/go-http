package http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	stdhttp "net/http"
	"strings"
)

type Response struct {
	Headers stdhttp.Header
	Status  int
	Body    io.ReadCloser
	bytes   []byte
}

func (r *Response) Bytes() ([]byte, error) {
	if r.bytes != nil {
		return r.bytes, nil
	}

	if r.Body == nil {
		return nil, nil
	}

	b, err := io.ReadAll(r.Body)
	if err != nil {
		return b, err
	}

	r.bytes = b
	if err := r.Body.Close(); err != nil {
		return b, err
	}

	r.Body = io.NopCloser(bytes.NewReader(b))
	return b, nil
}

// ResponseOption is the option type for responses
type ResponseOption interface {
	ProcessResponse(*Response) error
}

func (j jsonOption) ProcessResponse(resp *Response) error {
	ct := resp.Headers.Get("Content-Type")
	if ct != "" && !strings.HasPrefix(ct, "application/json") {
		return fmt.Errorf("invalid Content-Type header, expected 'application/json', got %s", ct)
	}
	b, err := resp.Bytes()
	if err != nil {
		return err
	}
	return json.Unmarshal(b, j.v)
}

func (resp *Response) applyOptions(options ...ResponseOption) error {
	for _, opt := range options {
		if err := opt.ProcessResponse(resp); err != nil {
			return err
		}
	}
	return nil
}
