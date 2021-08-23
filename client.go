package http

import (
	stdhttp "net/http"

	copy "github.com/mitchellh/copystructure"
)

type Client struct {
	baseClient *stdhttp.Client

	PreRequestMiddlewares  []RequestOption
	PostRequestMiddlewares []RequestOption

	PreResponseMiddlewares  []ResponseOption
	PostResponseMiddlewares []ResponseOption
}

// NewClient creates a new HTTP Client with the given options
func NewClient(options ...ClientOption) *Client {
	c := &Client{}
	c.Apply(options...)
	return c
}

func (c *Client) copy() *Client {
	c1 := new(Client)

	c1.baseClient = copy.Must(copy.Copy(c.baseClient)).(*stdhttp.Client)
	c1.PreRequestMiddlewares = copy.Must(copy.Copy(c.PreRequestMiddlewares)).([]RequestOption)
	c1.PostRequestMiddlewares = copy.Must(copy.Copy(c.PostRequestMiddlewares)).([]RequestOption)
	c1.PreResponseMiddlewares = copy.Must(copy.Copy(c.PreResponseMiddlewares)).([]ResponseOption)
	c1.PostResponseMiddlewares = copy.Must(copy.Copy(c.PostResponseMiddlewares)).([]ResponseOption)

	return c1
}

// With clones the client, applying the new options ontop
func (c *Client) With(options ...ClientOption) *Client {
	c1 := c.copy()
	c1.Apply(options...)
	return c1
}

// With clones the client, applying the new options ontop
func (c *Client) Apply(options ...ClientOption) {
	for _, opt := range options {
		opt.ModifyClient(c)
	}
}

// Get creates a new HTTP Get Request with the options provided
func (c *Client) Get(options ...RequestOption) *Request {
	return c.NewRequest(Get, options...)
}

// Post creates a new HTTP Post Request with the options provided
func (c *Client) Post(options ...RequestOption) *Request {
	return c.NewRequest(Post, options...)
}

// Put creates a new HTTP Put Request with the options provided
func (c *Client) Put(options ...RequestOption) *Request {
	return c.NewRequest(Put, options...)
}

// Delete creates a new HTTP Delete Request with the options provided
func (c *Client) Delete(options ...RequestOption) *Request {
	return c.NewRequest(Delete, options...)
}

// NewRequest creates a new HTTP Request with the method and options provided
func (c *Client) NewRequest(method Method, options ...RequestOption) *Request {
	req := Request{
		Client: c,
		Method: method,
	}

	req.err = req.applyOptions(c.PreRequestMiddlewares...)
	if req.err != nil {
		return &req
	}

	req.err = req.applyOptions(options...)

	return &req
}

func (c *Client) BaseClient() *stdhttp.Client {
	if c.baseClient == nil {
		return stdhttp.DefaultClient
	}
	return c.baseClient
}
