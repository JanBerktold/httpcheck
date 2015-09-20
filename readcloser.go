package httpcheck

import "io"

type ReadCloser struct {
	rd io.Reader
}

func NewReadCloser(rd io.Reader) *ReadCloser {
	return &ReadCloser{rd}
}

func (rc *ReadCloser) Read(p []byte) (n int, err error) {
	return rc.rd.Read(p)
}

func (rc *ReadCloser) Close() error {
	return nil
}
