package http

import (
	"fmt"
	stdhttp "net/http"
	"net/url"

	copy "github.com/mitchellh/copystructure"
)

type Client struct {
	baseURL     *url.URL
	baseHeaders stdhttp.Header

	transport           stdhttp.RoundTripper
	requestMiddlewares  []RequestOption
	responseMiddlewares []ResponseOption

	err error
}

// NewClient creates a new HTTP Client with the given options
func NewClient(options ...ClientOption) *Client {
	c := &Client{}
	c.err = c.applyOptions(options...)
	return c
}

func (c *Client) applyOptions(options ...ClientOption) error {
	for _, opt := range options {
		if err := opt.ModifyClient(c); err != nil {
			return err
		}
	}
	return nil
}

func (c *Client) copy() *Client {
	c1 := new(Client)

	c1.baseURL = copy.Must(copy.Copy(c.baseURL)).(*url.URL)
	c1.baseHeaders = copy.Must(copy.Copy(c.baseHeaders)).(stdhttp.Header)
	c1.transport = copy.Must(copy.Copy(c.transport)).(stdhttp.RoundTripper)
	c1.requestMiddlewares = copy.Must(copy.Copy(c.requestMiddlewares)).([]RequestOption)
	c1.responseMiddlewares = copy.Must(copy.Copy(c.responseMiddlewares)).([]ResponseOption)
	c1.err = c.err

	return c1
}

// With clones the client, applying the new options ontop
func (c *Client) With(options ...ClientOption) *Client {
	if c.err != nil {
		return c
	}

	c1 := c.copy()
	c1.err = c1.applyOptions(options...)
	return c1
}

// NewRequest creates a new HTTP Request with the method and options provided
func (c *Client) NewRequest(method Method, options ...RequestOption) *Request {
	if c.err != nil {
		return &Request{err: fmt.Errorf("client error: %w", c.err)}
	}

	req := Request{
		client:  c,
		method:  method,
		url:     c.baseURL,
		headers: c.baseHeaders,
	}

	req.err = req.applyOptions(options...)

	return &req
}

func (c *Client) roundTripper() stdhttp.RoundTripper {
	if c.transport == nil {
		return stdhttp.DefaultTransport
	}
	return c.transport
}

// ClientOption is the option type for clients
type ClientOption interface {
	ModifyClient(*Client) error
}

func (u URLOption) ModifyClient(c *Client) error {
	c.baseURL = u.url
	return nil
}

func (u URLStringOption) ModifyClient(c *Client) error {
	url, err := url.Parse(u.url)
	if err != nil {
		return err
	}
	return URL(url).ModifyClient(c)
}

func (h HeaderOption) ModifyClient(c *Client) error {
	if c.baseHeaders == nil {
		c.baseHeaders = h.headers
	} else {
		for k, vs := range h.headers {
			for _, v := range vs {
				c.baseHeaders.Add(k, v)
			}
		}
	}
	return nil
}

type RequestOptions struct {
	options []RequestOption
}

func RequestMiddlewares(options ...RequestOption) RequestOptions {
	return RequestOptions{options}
}

func (r RequestOptions) ModifyClient(c *Client) error {
	c.requestMiddlewares = append(c.requestMiddlewares, r.options...)
	return nil
}

type ResponseOptions struct {
	options []ResponseOption
}

func ResponseMiddlewares(options ...ResponseOption) ResponseOptions {
	return ResponseOptions{options}
}

func (r ResponseOptions) ModifyClient(c *Client) error {
	c.responseMiddlewares = append(c.responseMiddlewares, r.options...)
	return nil
}

type TransportOption struct {
	transport stdhttp.RoundTripper
}

func Transport(rt stdhttp.RoundTripper) TransportOption {
	return TransportOption{rt}
}

func (t TransportOption) ModifyClient(c *Client) error {
	c.transport = t.transport
	return nil
}
