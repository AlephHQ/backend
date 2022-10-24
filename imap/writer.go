package imap

import (
	"bufio"
	"io"
)

const crlf = "\r\n"

type Writer struct {
	*bufio.Writer
}

func (w *Writer) WriteString(s string) (int, error) {
	return w.Writer.WriteString(s + crlf)
}

func NewWriter(w io.Writer) *Writer {
	wr := &Writer{}
	wr.Writer = bufio.NewWriter(w)

	return wr
}
