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

type JSONOption struct {
	v interface{}
}

func (j JSONOption) ModifyRequest(r *Request) error {
	b := bytes.NewBuffer(nil)
	if err := json.NewEncoder(b).Encode(j.v); err != nil {
		return fmt.Errorf("cannot encode request body: %w", err)
	}
	return r.applyOptions(Body(b), Header("Content-Type", "application/json"))
}

// JSON is an option to add a JSON Body to a request or to expect a JSON Body in a response
func JSON(v interface{}) JSONOption {
	return JSONOption{v}
}

type HeaderOption struct {
	key, value string
}

// Header is an option to add a HTTP header to a request
func Header(key, value string) HeaderOption {
	return HeaderOption{key, value}
}

func (h HeaderOption) ModifyRequest(r *Request) error {
	if r.headers == nil {
		r.headers = stdhttp.Header{}
	}
	r.headers.Add(h.key, h.value)
	return nil
}

type BodyOption struct {
	r io.Reader
}

func (b BodyOption) ModifyRequest(r *Request) error {
	rc, ok := b.r.(io.ReadCloser)
	if !ok && b.r != nil {
		rc = io.NopCloser(b.r)
	}
	r.body = rc
	return nil
}

// Body is an option to add a body to a request
func Body(r io.Reader) BodyOption {
	return BodyOption{r}
}

type PathOption struct {
	segments []string
}

// Path is an option to add a path onto the base url of the request
func Path(pathSegments ...string) PathOption {
	return PathOption{pathSegments}
}

func (p PathOption) ModifyRequest(r *Request) error {
	if r.url == nil {
		return fmt.Errorf("cannot use path option because there's no base url")
	}
	r.url.Path = path.Join(append([]string{r.url.Path}, p.segments...)...)
	return nil
}

type QueryOption struct {
	values url.Values
}

// Query is an option to add query parameters onto a request
func Query(values url.Values) QueryOption {
	return QueryOption{values}
}

// QueryValue is an option to add a single query parameter onto a request
func QueryValue(key, value string) QueryOption {
	return QueryOption{values: url.Values{
		key: []string{value},
	}}
}

func (q QueryOption) ModifyRequest(r *Request) error {
	query := r.url.Query()
	for k, vs := range q.values {
		for _, v := range vs {
			query.Add(k, v)
		}
	}
	r.url.RawQuery = query.Encode()
	return nil
}

type URLOption struct {
	url *url.URL
}

// URL is an option to set the url of the request
func URL(url *url.URL) URLOption {
	return URLOption{url}
}

func (u URLOption) ModifyRequest(r *Request) error {
	r.url = u.url
	return nil
}

type URLStringOption struct {
	url string
}

// URL is an option to parse and set the url of the request
func URLString(url string) URLStringOption {
	return URLStringOption{url}
}

func (u URLStringOption) ModifyRequest(r *Request) error {
	url, err := url.Parse(u.url)
	if err != nil {
		return err
	}
	r.url = url
	return nil
}
