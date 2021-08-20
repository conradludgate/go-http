package http_test

import (
	"context"
	stdhttp "net/http"
	"net/url"
	"testing"

	"github.com/conradludgate/go-http"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func URL(t *testing.T, u string) url.URL {
	t.Helper()
	_u, err := url.Parse(u)
	require.NoError(t, err)
	return *_u
}

type FooBar struct {
	Foo string
	Bar int
}

func TestReq(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	client := http.NewClient().WithBaseURL(URL(t, "https://example.com/api"))
	ctx := context.Background()

	httpmock.RegisterResponder("GET", "https://example.com/api/foo/bar",
		RespondWith(JSON(object{
			"Foo": "something",
			"Bar": 234,
			"Baz": "12h",
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

	resp, err := client.NewRequest(http.Get, http.Path("foo", "bar"),
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
