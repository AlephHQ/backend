package conn

import (
	"bufio"
	"io"
)

type Reader struct {
	*bufio.Reader
}

func NewReader(r io.Reader) *Reader {
	reader := &Reader{}
	reader.Reader = bufio.NewReader(r)

	return reader
}
