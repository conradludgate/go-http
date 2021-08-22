package http

import (
	"bytes"
	"fmt"
	"io"
	stdhttp "net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResponse_Bytes_Reuse(t *testing.T) {
	bodyBytes := []byte("foobarbaz")
	body := bytes.NewReader(bodyBytes)

	resp := Response{
		body: newResponseReader(io.NopCloser(body)),
	}
	b, err := io.ReadAll(&resp)
	require.Nil(t, err)
	assert.Equal(t, bodyBytes, b)

	resp.Reset()

	b, err = io.ReadAll(&resp)
	require.Nil(t, err)
	assert.Equal(t, bodyBytes, b)
}

type errReader struct {
	err error
}

func (e errReader) Read(p []byte) (int, error) {
	return 0, e.err
}

func TestResponse_Bytes_ReadError(t *testing.T) {
	resp := Response{
		body: newResponseReader(io.NopCloser(errReader{fmt.Errorf("error reading data")})),
	}
	b, err := io.ReadAll(&resp)
	assert.EqualError(t, err, "error reading data")
	assert.Empty(t, b)
}

type errCloser struct {
	r   io.Reader
	err error
}

func (e errCloser) Read(p []byte) (int, error) {
	if e.r == nil {
		return 0, io.EOF
	}
	n, err := e.r.Read(p)
	if err != nil {
		e.r = nil
	}
	return n, err
}

func (e errCloser) Close() error {
	return e.err
}

func TestResponse_Bytes_CloseError(t *testing.T) {
	bodyBytes := []byte("foobarbaz")
	body := bytes.NewReader(bodyBytes)

	resp := Response{
		body: newResponseReader(errCloser{
			body,
			fmt.Errorf("error closing conn"),
		}),
	}
	err := resp.Close()
	assert.EqualError(t, err, "error closing conn")
}

func TestResponse_JSON_Error_ContentType(t *testing.T) {
	bodyBytes := []byte(`{"Foo":"bar"}`)
	body := bytes.NewReader(bodyBytes)

	resp := Response{
		body: newResponseReader(io.NopCloser(body)),
		Headers: stdhttp.Header{
			"Content-Type": []string{"text/plain"},
		},
	}

	err := JSON(nil).ProcessResponse(&resp)
	assert.EqualError(t, err, "invalid Content-Type header, expected 'application/json', got text/plain")
}

func TestResponse_JSON_Error_Unmarshal(t *testing.T) {
	bodyBytes := []byte(`{"Foo":"bar"}`)
	body := bytes.NewReader(bodyBytes)

	resp := Response{
		body: newResponseReader(io.NopCloser(body)),
		Headers: stdhttp.Header{
			"Content-Type": []string{"application/json"},
		},
	}

	output := new(struct{ Foo int })
	err := JSON(output).ProcessResponse(&resp)
	assert.EqualError(t, err, "json: cannot unmarshal string into Go struct field .Foo of type int")
	assert.Zero(t, *output)
}
