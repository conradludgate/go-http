# go-http
Simple/Extensible http request library for go

[![GoDoc](https://godoc.org/github.com/conradludgate/go-http?status.svg)](http://godoc.org/github.com/zmb3/spotify)
![latest version](https://img.shields.io/github/v/tag/conradludgate/go-http?label=version)
[![code coverage](https://img.shields.io/codecov/c/gh/conradludgate/go-http)](https://app.codecov.io/gh/conradludgate/go-http/)

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
req := client.NewRequest(http.Get, http.Path("v1", "healthz"))
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
