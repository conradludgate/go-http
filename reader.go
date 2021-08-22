package http

import (
	"io"
)

type buffer struct {
	buffer []byte
	offset int
}

func (r *buffer) Len() int {
	return len(r.buffer) - r.offset
}

func (r *buffer) Read(p []byte) (n int, err error) {
	if remaining := r.Len(); remaining < len(p) {
		n = copy(p[:remaining], r.buffer[r.offset:])
		err = io.EOF
	} else {
		n = copy(p, r.buffer[r.offset:])
	}
	r.offset += n
	return
}

func (r *buffer) Write(p []byte) (n int, err error) {
	r.buffer = append(r.buffer, p...)
	r.offset = len(r.buffer)
	return len(p), nil
}

func (r *buffer) Reset() {
	r.offset = 0
}

type responseReader struct {
	reader io.ReadCloser
	buffer *buffer
}

func newResponseReader(r io.ReadCloser) *responseReader {
	return &responseReader{
		reader: r,
		buffer: new(buffer),
	}
}

func (r *responseReader) Read(p []byte) (n int, err error) {
	if r.buffer.Len() == 0 && r.reader != nil {
		n, err = io.TeeReader(r.reader, r.buffer).Read(p)
		if err != nil {
			_ = r.reader.Close() // try close on error (most likely EOF). Ignoring read close errors...
			r.reader = nil
		}
		return
	} else {
		return r.buffer.Read(p)
	}
}

func (r *responseReader) Reset() {
	r.buffer.Reset()
}

func (r *responseReader) Close() error {
	if r.reader != nil {
		return r.reader.Close()
	}
	return nil
}
