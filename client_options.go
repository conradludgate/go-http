package http

import (
	stdhttp "net/http"
)

// ClientOption is the option type for clients
type ClientOption interface {
	ModifyClient(*Client)
}

func (u URLOption) ModifyClient(c *Client) {
	PreRequestMiddlewares(u).ModifyClient(c)
}

func (u URLStringOption) ModifyClient(c *Client) {
	PreRequestMiddlewares(u).ModifyClient(c)
}

func (h HeaderOption) ModifyClient(c *Client) {
	PreRequestMiddlewares(h).ModifyClient(c)
}

type PreRequestOptions struct {
	options []RequestOption
}

func PreRequestMiddlewares(options ...RequestOption) PreRequestOptions {
	return PreRequestOptions{options}
}

func (r PreRequestOptions) ModifyClient(c *Client) {
	c.PreRequestMiddlewares = append(c.PreRequestMiddlewares, r.options...)
}

type PostRequestOptions struct {
	options []RequestOption
}

func PostRequestMiddlewares(options ...RequestOption) PostRequestOptions {
	return PostRequestOptions{options}
}

func (r PostRequestOptions) ModifyClient(c *Client) {
	c.PostRequestMiddlewares = append(c.PostRequestMiddlewares, r.options...)
}

type PreResponseOptions struct {
	options []ResponseOption
}

func PreResponseMiddlewares(options ...ResponseOption) PreResponseOptions {
	return PreResponseOptions{options}
}

func (r PreResponseOptions) ModifyClient(c *Client) {
	c.PreResponseMiddlewares = append(c.PreResponseMiddlewares, r.options...)
}

type PostResponseOptions struct {
	options []ResponseOption
}

func PostResponseMiddlewares(options ...ResponseOption) PostResponseOptions {
	return PostResponseOptions{options}
}

func (r PostResponseOptions) ModifyClient(c *Client) {
	c.PostResponseMiddlewares = append(c.PostResponseMiddlewares, r.options...)
}

type BaseClientOption struct {
	client *stdhttp.Client
}

func BaseClient(client *stdhttp.Client) BaseClientOption {
	return BaseClientOption{client}
}

func (co BaseClientOption) ModifyClient(c *Client) {
	c.baseClient = co.client
}
