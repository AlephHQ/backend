package conn

import (
	"bufio"
	"io"
)

const crlf = "\r\n"

type IMAPWriter struct {
	*bufio.Writer
}

func NewIMAPWriter(w io.Writer) *IMAPWriter {
	wr := &IMAPWriter{}
	wr.Writer = bufio.NewWriter(w)

	return wr
}

func (w *IMAPWriter) writeCommand(cmd string) error {
	_, err := w.Writer.WriteString(cmd + crlf)
	if err != nil {
		return err
	}

	return w.Writer.Flush()
}
