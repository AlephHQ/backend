package imap

import (
	"bufio"
	"crypto/rand"
	"encoding/base64"
	"io"
	"log"
)

const crlf = "\r\n"

func getTag() string {
	b := make([]byte, 7)
	_, err := rand.Read(b)
	if err != nil {
		log.Panic(err)
	}

	return base64.RawURLEncoding.EncodeToString(b)
}

type Writer struct {
	*bufio.Writer
}

func NewWriter(w io.Writer) *Writer {
	wr := &Writer{}
	wr.Writer = bufio.NewWriter(w)

	return wr
}

func (w *Writer) WriteString(s string) error {
	cmd := getTag() + " " + s
	log.Println(cmd)
	_, err := w.Writer.WriteString(cmd + crlf)
	if err != nil {
		return err
	}

	return w.Writer.Flush()
}
