package http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	stdhttp "net/http"
	"net/url"
	"path"
)

// RequestOption is the option type for requests
type RequestOption interface {
	ModifyRequest(*Request) error
}

func (req *Request) applyOptions(options ...RequestOption) error {
	for _, opt := range options {
		if err := opt.ModifyRequest(req); err != nil {
			return err
		}
	}
	return nil
}

type jsonOption struct {
	v interface{}
}

func (j jsonOption) ModifyRequest(r *Request) error {
	b := bytes.NewBuffer(nil)
	if err := json.NewEncoder(b).Encode(j.v); err != nil {
		return fmt.Errorf("cannot encode request body: %w", err)
	}
	return r.applyOptions(Body(b), Header("Content-Type", "application/json"))
}

// JSON is an option to add a JSON Body to a request or to expect a JSON Body in a response
func JSON(v interface{}) jsonOption {
	return jsonOption{v}
}

type headerOption struct {
	key, value string
}

// Header is an option to add a HTTP header to a request
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

// Body is an option to add a body to a request
func Body(r io.Reader) bodyOption {
	return bodyOption{r}
}

type pathOption []string

// Path is an option to add a path onto the base url of the request
func Path(pathSegments ...string) pathOption {
	return pathOption(pathSegments)
}

func (p pathOption) ModifyRequest(r *Request) error {
	if r.url == nil {
		return fmt.Errorf("cannot use path option because there's no base url")
	}
	r.url.Path = path.Join(append([]string{r.url.Path}, p...)...)
	return nil
}

type urlOption struct {
	url *url.URL
}

// URL is an option to set the url of the request
func URL(url *url.URL) urlOption {
	return urlOption{url}
}

func (u urlOption) ModifyRequest(r *Request) error {
	r.url = u.url
	return nil
}

type urlStringOption struct {
	url string
}

// URL is an option to parse and set the url of the request
func URLString(url string) urlStringOption {
	return urlStringOption{url}
}

func (u urlStringOption) ModifyRequest(r *Request) error {
	url, err := url.Parse(u.url)
	if err != nil {
		return err
	}
	r.url = url
	return nil
}
