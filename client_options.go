package http

import (
	stdhttp "net/http"
	"net/url"
)

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
