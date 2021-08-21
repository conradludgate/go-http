package http_test

import (
	"context"
	"fmt"
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

	client := http.NewClient(http.URLString("https://example.com/api"))
	ctx := context.Background()

	httpmock.RegisterResponder("GET", "https://example.com/api/foo/bar",
		RespondWith(JSON(object{
			"Foo": "something",
			"Bar": 234,
		}),
			VerifyJSONBody(object{
				"Foo": "foo",
				"Bar": 123,
			}),
		))

	reqBody := FooBar{
		Foo: "foo",
		Bar: 123,
	}
	respBody := FooBar{}

	resp, err := client.NewRequest(http.Get,
		http.Path("foo", "bar"),
		http.JSON(reqBody),
	).Send(ctx,
		http.JSON(&respBody),
	)

	require.NoError(t, err)
	assert.Equal(t, &http.Response{
		Status: 200,
		Headers: stdhttp.Header{
			"Content-Type": []string{"application/json"},
		},
	}, resp)
	assert.Equal(t, FooBar{
		Foo: "something",
		Bar: 234,
	}, respBody)
}

func ExampleJSON() {
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

	// make GET request to https://hacker-news.firebaseio.com/v0/item/8863
	req := client.NewRequest(http.Get,
		http.Path("v0", "item", "8863"),
	)

	respBody := new(HNItem)
	resp, err := req.Send(ctx, http.JSON(respBody))
	if err != nil {
		panic(err)
	}

	// Outputs: 200: My YC app: Dropbox - Throw away your USB drive
	fmt.Printf("%d: %s", resp.Status, respBody.Title)
}
