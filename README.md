# go-http
Opinionated, convenient and extensible http request library for go

[![GoDoc](https://godoc.org/github.com/conradludgate/go-http?status.svg)](http://godoc.org/github.com/conradludgate/go-http)
![latest version](https://img.shields.io/github/v/tag/conradludgate/go-http?label=version)
[![code coverage](https://img.shields.io/codecov/c/gh/conradludgate/go-http)](https://app.codecov.io/gh/conradludgate/go-http/)

## Why

The Go standard library is great and provides an amazing http package. This is not intended to be a hard replacement, but an expansion.

Let's say you're wrapping an API and you're sending a `POST` request with a `JSON` body, getting back some `JSON` data in the response.
Using the std lib, that might look something like

```go
import (
    "encoding/json"
    "fmt"
    "net/http"
)

type APIClient struct {
    baseURL    string
    httpClient *http.Client
}

func (c *APIClient) SendSomeData(ctx context.Context, req SomeData) (*SomeResponse, error) {
    url = c.baseURL + "/some/path"

    b, err := json.Marshal(req)
    if err != nil {
        return nil, fmt.Errorf("error marshaling json body: %w", err)
    }

    body := bytes.NewReader(b)
    req, err := http.NewRequestFromContext(ctx, http.MethodPost, url, body)
    if err != nil {
        return nil, fmt.Errorf("error creating request object: %w", err)
    }

    resp, err := c.httpClient.Do(req)
    if err != nil {
        return nil, fmt.Errorf("error sending http request: %w", err)
    }

    defer resp.Body.Close()
    var respBody SomeResponse
    if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
        return nil, fmt.Errorf("error decoding http response: %w", err)
    }

    return &respBody, nil
}
```

That's quite a bit of code. How about instead we do something like this

```go
import (
    "encoding/json"
    "fmt"

    "github.com/conradludgate/go-http"
)

type APIClient struct {
    httpClient *http.Client
}

func (c *APIClient) SendSomeData(ctx context.Context, req SomeData) (*SomeResponse, error) {
    var respBody SomeResponse

    _, err := c.httpClient.Post(
        http.Path("some", "path"),
        http.JSON(req)
    ).Send(ctx, http.JSON(&respBody))

    if err != nil {
        return nil, err
    }

    return &respBody, nil
}
```

That's a lot simpler. But how? Let's break it down.

First, how to we make the client?

```go
apiClient := &APIClient{
    // Create a new http client
    // It can accept many options, in this case we're setting
    // a base URL from a string
    http.NewClient(http.URLString("https://api.example.com/v1"))
}
```

How does the request work?

```go
// Using the http client, make a post request
c.httpClient.Post(
    // relative to the base url provided, make the request to `/some/path`.
    // eg `https://api.example.com/v1/some/path`
    http.Path("some", "path"),
    // Then add the `req` object as a JSON body to the request
    http.JSON(req),
).
// send the request with the provided context
Send(ctx,
    // Check the response has an `application/json` content type
    // and then decode the body into `respBody`
    http.JSON(&respBody),
)
```

## Installation

`go get github.com/conradludgate/go-http`

## Usage

This library consists of a few basic parts.

### Client

Clients are what make the HTTP requests. They are very simple to make

```go
client := http.NewClient()
```

The `NewClient` function accepts some options to extend it.

For example:

```go
client := http.NewClient(
    // Set the base url of the client
    http.URLString("https://example.com/api/"),
    // Set a header to be sent with every request
    http.AddHeader("Authorization", "Bearer ABC"),
)
```

You can create a temporary client scope using the `With` function

```go
client := http.NewClient(...)

// Creates a copy of client with a new header
// The original client is unaffected
client2 := client.With(
    http.AddHeader("User-Agent", "googlebot"),
)
```

### Requests

Once you have a client, you can create requests

```go
// Create a GET request object to <base_url>/v1/healthz
req := client.Get(http.Path("v1", "healthz"))
```

You can add a lot of options here too:

```go
// Create a POST request object
req := client.Post(
    // to https://example.com/api/v1/healthz
    http.URLString("https://example.com/api/"),
    http.Path("v1", "healthz"),
    // With a JSON body
    http.JSON(someData),
    // And a User-Agent header
    http.AddHeader("User-Agent", "googlebot"),
)
```

### Responses

Once you have your request object, you can `Send` it

```go
resp, err := req.Send(context.Background())
```

A repeating story... `Send` also accepts some options

```go
// any type that deserializes JSON
respBody := make(map[string]interface{})

resp, err := req.Send(context.Background(),
    http.JSON(&respBody)
)
```

## Examples

### Simple usage

```go
// Set the base url for the client
client := http.NewClient(http.URLString("https://hacker-news.firebaseio.com/"))

respBody := make(map[string]interface{})

// make GET request
resp, err := client.Get(
    // to https://hacker-news.firebaseio.com/v0/item/8863.json
    http.Path("v0", "item", "8863.json"),
).Send(context.Background(),
    // deserialising the json response into respBody
    http.JSON(&respBody),
)
if err != nil {
    panic(err)
}

// Output: 200: My YC app: Dropbox - Throw away your USB drive
fmt.Printf("%d: %s", resp.Status, respBody["title"])
```

### Complex usage

```go
type RequestBody struct {
    Foo string
    Bar int64
}
type ResponseBody struct {
    URL  string      `json:"url"`
    JSON RequestBody `json:"json"`
}

reqBody := RequestBody{
    Foo: "Hello World",
    Bar: 1234,
}
respBody := new(ResponseBody)

client := http.NewClient()
resp, err := client.NewRequest(http.Post,
    // Send post request to https://httpbin.org/anything/test1
    http.URLString("https://httpbin.org/anything"),
    http.Path("test1"),
    // Including a header
    http.AddHeader("X-Example", "wow"),
    // Sending json from reqBody
    http.JSON(reqBody),
).Send(context.Background(),
    // Receiving json response into respBody
    http.JSON(respBody),
)
if err != nil {
    panic(err)
}

// Output: 200: https://httpbin.org/anything/test1 {Hello World 1234}
fmt.Printf("%d: %s %v", resp.Status, respBody.URL, respBody.JSON)
```

### Errors

Most Go HTTP clients return errors when making request objects as well as sending them.

I usually find it annoying to make a request object, check the error,
send the request, check the error and in both cases just bubbling it up wrapped with some context.

```go
req, err := http.NewRequest(http.MethodGet, ":invalid_url")
if err != nil {
    return nil, fmt.Errorf("request error: ", err)
}
resp, err := client.Do(req)
if err != nil {
    return nil, fmt.Errorf("response error: ", err)
}
```

This library does that for you in one go.

```go
_, err := client.Get(http.URLString(":invalid_url")).Send(context.Background())

// Output: request error: parse \":invalid-url\": missing protocol scheme
fmt.Println(err)
```

If you want to, you can extract the error after creating the request

```go
req := client.Get(http.URLString(":invalid_url"))

// Output: request error: parse \":invalid-url\": missing protocol scheme
fmt.Println(req.Error())
```
