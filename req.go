package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	stdhttp "net/http"
	"net/url"
	"path"
)

type jsonOption struct {
	v interface{}
}

func (j jsonOption) ModifyRequest(r *Request) error {
	b := bytes.NewBuffer(nil)
	if err := json.NewEncoder(b).Encode(j.v); err != nil {
		return fmt.Errorf("cannot encode request body: %w", err)
	}
	bodyOption{b}.ModifyRequest(r)
	return headerOption{"Content-Type", "application/json"}.ModifyRequest(r)
}

func JSON(v interface{}) jsonOption {
	return jsonOption{v}
}

type headerOption struct {
	key, value string
}

func Header(key, value string) headerOption {
	return headerOption{key, value}
}

func (h headerOption) ModifyRequest(r *Request) error {
	if r.headers == nil {
		r.headers = stdhttp.Header{}
	}
	r.headers.Add(h.key, h.value)
	return nil
}

type bodyOption struct {
	r io.Reader
}

func (b bodyOption) ModifyRequest(r *Request) error {
	rc, ok := b.r.(io.ReadCloser)
	if !ok && b.r != nil {
		rc = io.NopCloser(b.r)
	}
	r.body = rc
	return nil
}

func Body(r io.Reader) bodyOption {
	return bodyOption{r}
}

type RequestOption interface {
	ModifyRequest(*Request) error
}

type PathSegments []string

func Path(pathSegments ...string) PathSegments {
	return PathSegments(pathSegments)
}

func (p PathSegments) joinTo(url *url.URL) {
	url.Path = path.Join(append([]string{url.Path}, p...)...)
}

type Client struct {
	baseURL     url.URL
	baseHeaders stdhttp.Header
}

func NewClient() Client {
	return Client{}
}

func (c Client) WithBaseURL(url url.URL) Client {
	c.baseURL = url
	return c
}

func (c Client) NewRequest(method Method, path PathSegments, options ...RequestOption) Request {
	req := Request{
		transport: stdhttp.DefaultTransport,
		method:    method,
		url:       c.baseURL,
		headers:   c.baseHeaders,
	}

	path.joinTo(&req.url)

	for _, opt := range options {
		if err := opt.ModifyRequest(&req); err != nil {
			req.errors = append(req.errors, err)
		}
	}

	return req
}

type Method string

const (
	Get  Method = "GET"
	Post Method = "POST"
)

type Request struct {
	transport stdhttp.RoundTripper

	method  Method
	url     url.URL
	body    io.ReadCloser
	headers stdhttp.Header

	errors []error
}

type ErrorGroup []error

func (g ErrorGroup) Error() string {
	return g[0].Error()
}

func (r *Request) Error() error {
	if len(r.errors) == 0 {
		return nil
	}
	return ErrorGroup(r.errors)
}

func (r Request) Send(ctx context.Context, options ...ResponseOption) (*Response, error) {
	if err := r.Error(); err != nil {
		return nil, err
	}

	req := &stdhttp.Request{
		Method:     string(r.method),
		URL:        &r.url,
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Header:     r.headers,
		Body:       r.body,
		Host:       r.url.Host,
	}

	stdresp, err := r.transport.RoundTrip(req.WithContext(ctx))
	if err != nil {
		return nil, err
	}

	resp := Response{
		Headers: stdresp.Header,
		Status:  stdresp.StatusCode,
		Body:    stdresp.Body,
	}

	for _, opt := range options {
		if err := opt.ProcessResponse(&resp); err != nil {
			return nil, err
		}
	}

	return &resp, nil
}

type Response struct {
	Headers stdhttp.Header
	Status  int
	Body    io.ReadCloser
}

type ResponseOption interface {
	ProcessResponse(*Response) error
}

func (j jsonOption) ProcessResponse(resp *Response) error {
	ct := resp.Headers.Get("Content-Type")
	if ct != "" && ct != "application/json" {
		return fmt.Errorf("invalid Content-Type header, expected 'application/json', got %s", ct)
	}
	if resp.Body == nil {
		return fmt.Errorf("response has no body to decode")
	}

	defer func() {
		resp.Body.Close()
		resp.Body = nil
	}()
	return json.NewDecoder(resp.Body).Decode(j.v)
}
