package imap

import (
	"bufio"
	"io"
	"log"
)

const crlf = "\r\n"

type Writer struct {
	*bufio.Writer
}

func NewWriter(w io.Writer) *Writer {
	wr := &Writer{}
	wr.Writer = bufio.NewWriter(w)

	return wr
}

func (w *Writer) WriteString(cmd string) error {
	log.Println(cmd)
	_, err := w.Writer.WriteString(cmd + crlf)
	if err != nil {
		return err
	}

	return w.Writer.Flush()
}
