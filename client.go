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

// Get creates a new HTTP Get Request with the options provided
func (c *Client) Get(options ...RequestOption) *Request {
	return c.NewRequest(Get, options...)
}

// Post creates a new HTTP Post Request with the options provided
func (c *Client) Post(options ...RequestOption) *Request {
	return c.NewRequest(Post, options...)
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
