package http_test

import (
	"context"
	"fmt"
	"io"
	stdhttp "net/http"
	"net/url"
	"testing"

	"github.com/conradludgate/go-http"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func URL(t *testing.T, u string) *url.URL {
	t.Helper()
	_u, err := url.Parse(u)
	require.NoError(t, err)
	return _u
}

type FooBar struct {
	Foo string
	Bar int
}

func TestReq(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("GET", "https://example.com/api/foo/bar",
		RespondWith(JSON(FooBar{
			Foo: "something",
			Bar: 234,
		}),
			VerifyJSONBody(FooBar{
				Foo: "foo",
				Bar: 123,
			}),
			VerifyHeader("foo", "bar", "baz"),
		))

	client := http.NewClient(
		http.URLString("https://example.com/api"),
		http.BaseClient(stdhttp.DefaultClient),
	)
	ctx := context.Background()

	reqBody := FooBar{
		Foo: "foo",
		Bar: 123,
	}
	respBody := FooBar{}

	resp, err := client.NewRequest(http.Get,
		http.Path("foo", "bar"),
		http.JSON(reqBody),
		http.AddHeader("foo", "bar", "baz"),
	).Send(ctx,
		http.JSON(&respBody),
	)

	require.NoError(t, err)
	assert.Equal(t, http.Status(200), resp.StatusCode)
	assert.Equal(t, stdhttp.Header{
		"Content-Type": []string{"application/json"},
	}, resp.Headers)
	assert.Equal(t, FooBar{
		Foo: "something",
		Bar: 234,
	}, respBody)
}

func ExampleRequest_Send() {
	type HNItem struct {
		ID         int64  `json:"id"`
		Type       string `json:"type"`
		Decendants int64  `json:"decendants"`

		By    string `json:"by"`
		Time  int64  `json:"time"`
		Score int64  `json:"score"`

		URL   string `json:"url"`
		Title string `json:"title"`

		Kids []int64 `json:"kids"`
	}

	ctx := context.Background()

	// Set the base url for the client
	client := http.NewClient(http.URLString("https://hacker-news.firebaseio.com/"))

	// make GET request to https://hacker-news.firebaseio.com/v0/item/8863.json
	req := client.Get(http.Path("v0", "item", "8863.json"))

	respBody := new(HNItem)
	resp, err := req.Send(ctx, http.JSON(respBody))
	if err != nil {
		panic(err)
	}

	fmt.Printf("%d: %s", resp.StatusCode, respBody.Title)

	// Output: 200: My YC app: Dropbox - Throw away your USB drive
}

func ExampleJSON() {
	type RequestBody struct {
		Foo string
		Bar int64
	}
	type ResponseBody struct {
		URL  string      `json:"url"`
		JSON RequestBody `json:"json"`
	}

	ctx := context.Background()

	reqBody := RequestBody{
		Foo: "Hello World",
		Bar: 1234,
	}
	respBody := new(ResponseBody)

	client := http.NewClient()
	resp, err := client.Post(
		http.URLString("https://httpbin.org/anything"),
		http.Path("test1"),
		http.JSON(reqBody),
	).Send(ctx,
		http.JSON(respBody),
	)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%d: %s %v", resp.StatusCode, respBody.URL, respBody.JSON)

	// Output: 200: https://httpbin.org/anything/test1 {Hello World 1234}
}

func ExampleParam() {
	type ResponseBody struct {
		Args map[string]string `json:"args"`
	}

	ctx := context.Background()

	respBody := new(ResponseBody)

	client := http.NewClient()
	resp, err := client.Get(
		http.URLString("https://httpbin.org/get"),
		http.Param("foo", "abc"),
		http.Params(url.Values{
			"bar": []string{"def"},
		}),
	).Send(ctx,
		http.JSON(respBody),
	)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%d: %s %s", resp.StatusCode, respBody.Args["foo"], respBody.Args["bar"])

	// Output: 200: abc def
}

func TestClientURLError(t *testing.T) {
	client := http.NewClient(http.URLString(":invalid-url"))
	ctx := context.Background()

	resp, err := client.NewRequest(http.Get, http.Path("foo", "bar")).Send(ctx)

	assert.EqualError(t, err, "request error: parse \":invalid-url\": missing protocol scheme")
	assert.Nil(t, resp)
}

func TestRequestURLError(t *testing.T) {
	client := http.NewClient()
	ctx := context.Background()

	resp, err := client.NewRequest(http.Get, http.URLString(":invalid-url")).Send(ctx)

	assert.EqualError(t, err, "request error: parse \":invalid-url\": missing protocol scheme")
	assert.Nil(t, resp)
}

func TestPathError(t *testing.T) {
	client := http.NewClient()
	ctx := context.Background()

	resp, err := client.NewRequest(http.Get, http.Path("foo", "bar")).Send(ctx)

	assert.EqualError(t, err, "request error: cannot use path option because there's no base url")
	assert.Nil(t, resp)

	resp, err = client.NewRequest(http.Get, http.URLString("relative-url"), http.Path("foo", "bar")).Send(ctx)

	assert.EqualError(t, err, "request error: cannot use path option because there's no base url")
	assert.Nil(t, resp)
}

func TestJSONEncodeError(t *testing.T) {
	client := http.NewClient()
	ctx := context.Background()

	resp, err := client.NewRequest(http.Get,
		http.URLString("http://example.com"), http.Path("foo", "bar"),
		http.JSON(make(chan bool)),
	).Send(ctx)

	assert.EqualError(t, err, "request error: cannot encode request body: json: unsupported type: chan bool")
	assert.Nil(t, resp)
}

func TestQuery(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	client := http.NewClient()
	ctx := context.Background()

	httpmock.RegisterResponder("GET", "https://example.com/api?foo=bar",
		httpmock.NewStringResponder(200, "correct"))

	resp, err := client.NewRequest(http.Get,
		http.URLString("https://example.com/api"),
		http.Param("foo", "bar"),
	).Send(ctx)

	require.Nil(t, err)
	assert.Equal(t, http.Status(200), resp.StatusCode)

	b, err := io.ReadAll(resp)
	require.Nil(t, err)
	assert.Equal(t, "correct", string(b))
}
