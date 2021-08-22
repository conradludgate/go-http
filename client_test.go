package http

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_With(t *testing.T) {
	client1 := NewClient(
		URLString("https://example.com"),
		AddHeader("foo", "bar", "baz"),
		RequestMiddlewares(JSON(nil)),
		ResponseMiddlewares(JSON(nil)),
		Transport(http.DefaultTransport),
	)

	client2 := client1.With(
		URLString("https://example.net"),
		AddHeader("abc", "bar", "baz"),
		RequestMiddlewares(AddHeader("foo", "bar", "baz")),
		ResponseMiddlewares(JSON(nil)),
		Transport(nil),
	)

	assert.Equal(t, "https://example.com", client1.baseURL.String())
	assert.Equal(t, http.Header{
		"foo": []string{"bar", "baz"},
	}, client1.baseHeaders)
	assert.Len(t, client1.requestMiddlewares, 1)
	assert.Len(t, client1.responseMiddlewares, 1)
	assert.Equal(t, client1.transport, http.DefaultTransport)

	assert.Equal(t, "https://example.net", client2.baseURL.String())
	assert.Equal(t, http.Header{
		"foo": []string{"bar", "baz"},
		"Abc": []string{"bar", "baz"},
	}, client2.baseHeaders)
	assert.Len(t, client2.requestMiddlewares, 2)
	assert.Len(t, client2.responseMiddlewares, 2)
	assert.Nil(t, client2.transport)
}

func TestClient_WithErr(t *testing.T) {
	client1 := NewClient(
		URLString(":invalid_url"),
	)

	require.EqualError(t, client1.err, "parse \":invalid_url\": missing protocol scheme")

	client2 := client1.With(
		URLString("https://example.com"),
	)

	assert.Equal(t, client1, client2)
}
