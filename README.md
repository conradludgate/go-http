# go-http
Simple/Extensible http request library for go

[![GoDoc](https://godoc.org/github.com/conradludgate/go-http?status.svg)](http://godoc.org/github.com/zmb3/spotify)
![latest version](https://img.shields.io/github/v/tag/conradludgate/go-http?label=version)
[![code coverage](https://img.shields.io/codecov/c/gh/conradludgate/go-http)](https://app.codecov.io/gh/conradludgate/go-http/)

## Examples

### Get Request

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

### Post Request

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
resp, err := client.Post(
    http.URLString("https://httpbin.org/anything"),
    http.Path("test1"),
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
