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
	Get    Method = "GET"
	Post   Method = "POST"
	Delete Method = "DELETE"
	Put    Method = "PUT"
)

type Request struct {
	Client *Client

	Method  Method
	URL     *url.URL
	Body    io.ReadCloser
	Headers stdhttp.Header

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

	if err := r.applyOptions(r.Client.PostRequestMiddlewares...); err != nil {
		return nil, err
	}

	req, err := stdhttp.NewRequestWithContext(ctx, string(r.Method), r.URL.String(), r.Body)
	if err != nil {
		panic(err)
	}
	if r.Headers != nil {
		req.Header = r.Headers
	}

	stdresp, err := r.Client.BaseClient().Do(req)
	if err != nil {
		return nil, err
	}

	resp := Response{
		Headers:    stdresp.Header,
		StatusCode: Status(stdresp.StatusCode),
		body:       newResponseReader(stdresp.Body),
	}

	if err := resp.applyOptions(r.Client.PreResponseMiddlewares...); err != nil {
		return &resp, err
	}

	if err := resp.applyOptions(options...); err != nil {
		return &resp, err
	}

	if err := resp.applyOptions(r.Client.PostResponseMiddlewares...); err != nil {
		return &resp, err
	}

	return &resp, nil
}
