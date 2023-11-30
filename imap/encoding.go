package imap

type Encoding string

const (
	Encoding7Bit           Encoding = "7bit"
	Encoding8Bit           Encoding = "8bit"
	EncodingQuotePrintable Encoding = "quoted-printable"
)
