package http

import (
	"context"
	"fmt"
	"io"
	stdhttp "net/http"
	"net/url"
)

type Method string

const (
	Get  Method = "GET"
	Post Method = "POST"
)

type Request struct {
	client *Client

	method  Method
	url     *url.URL
	body    io.ReadCloser
	headers stdhttp.Header

	err error
}

// Extract any errors out of the request that may have occured when building
// Done automatically when Sending but provided for extensibility
func (r *Request) Error() error {
	return r.err
}

// Send the HTTP Request, processing the response with the options provided
func (r *Request) Send(ctx context.Context, options ...ResponseOption) (*Response, error) {
	if r.err != nil {
		return nil, fmt.Errorf("request error: %w", r.err)
	}

	if err := r.applyOptions(r.client.requestMiddlewares...); err != nil {
		return nil, err
	}

	req := &stdhttp.Request{
		Method:     string(r.method),
		URL:        r.url,
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Header:     r.headers,
		Body:       r.body,
		Host:       r.url.Host,
	}

	if req.Header == nil {
		req.Header = stdhttp.Header{}
	}

	stdresp, err := r.client.roundTripper().RoundTrip(req.WithContext(ctx))
	if err != nil {
		return nil, err
	}

	resp := Response{
		Headers: stdresp.Header,
		Status:  stdresp.StatusCode,
		Body:    stdresp.Body,
	}

	if err := resp.applyOptions(append(r.client.responseMiddlewares, options...)...); err != nil {
		return nil, err
	}

	return &resp, nil
}
