package http

import (
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
	c1 := copy.Must(copy.Copy(c)).(*Client)
	if c1.transport == nil && c.transport != nil {
		c1.transport = copy.Must(copy.Copy(c.transport)).(stdhttp.RoundTripper)
	}
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
		return &Request{err: c.err}
	}

	req := Request{
		client:  c,
		method:  method,
		url:     c.baseURL,
		headers: c.baseHeaders,
	}
	req.applyOptions(options...)

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

func (u urlOption) ModifyClient(c *Client) error {
	c.baseURL = u.url
	return nil
}

func (u urlStringOption) ModifyClient(c *Client) error {
	url, err := url.Parse(u.url)
	if err != nil {
		return err
	}
	c.baseURL = url
	return nil
}