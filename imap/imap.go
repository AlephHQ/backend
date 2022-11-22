package imap

import (
	"regexp"
)

// RFC3501 Section 3
type ConnectionState int

const (
	ConnectingState       ConnectionState = 0
	NotAuthenticatedState ConnectionState = 1
	AuthenticatedState    ConnectionState = 2
	SelectedState         ConnectionState = 3
	LogoutState           ConnectionState = 4
	ConnectedState        ConnectionState = 5
)

// RFC3501 Section 7
type StatusResponse string

const (
	StatusResponseOK      StatusResponse = "OK"
	StatusResponseNO      StatusResponse = "NO"
	StatusResponseBYE     StatusResponse = "BYE"
	StatusResponseBAD     StatusResponse = "BAD"
	StatusResponsePREAUTH StatusResponse = "PREAUTH"
)

type StatusResponseCode string

const (
	StatusResponseCodeAlert          StatusResponseCode = "ALERT"
	StatusResponseCodeBadCharset     StatusResponseCode = "BADCHARSET"
	StatusResponseCodeCapability     StatusResponseCode = "CAPABILITY"
	StatusResponseCodeParse          StatusResponseCode = "PARSE"
	StatusResponseCodePermanentFlags StatusResponseCode = "PERMANENTFLAGS"
	StatusResponseCodeReadOnly       StatusResponseCode = "READ-ONLY"
	StatusResponseCodeReadWrite      StatusResponseCode = "READ-WRITE"
	StatusResponseCodeTryCreate      StatusResponseCode = "TRYCREATE"
	StatusResponseCodeUIDNext        StatusResponseCode = "UIDNEXT"
	StatusResponseCodeUIDValidity    StatusResponseCode = "UIDVALIDITY"
	StatusResponseCodeUnseen         StatusResponseCode = "UNSEEN"
)

type DataResponseCode string

const (
	DataResponseCodeFlags  DataResponseCode = "FLAGS"
	DataResponseCodeExists DataResponseCode = "EXISTS"
	DataResponseCodeRecent DataResponseCode = "RECENT"
)

type ResponseCode string

const (
	ResponseCodeFetch  ResponseCode = "FETCH"
	ResponseCodeSearch ResponseCode = "SEARCH"
)

type MessageAttribute string

const (
	MessageAttributeFlags        MessageAttribute = "FLAGS"
	MessageAttributeInternalDate MessageAttribute = "INTERNALDATE"
	MessageAttributeRFC822Size   MessageAttribute = "RFC822.SIZE"
	MessageAttributeEnvelope     MessageAttribute = "ENVELOPE"
	MessageAttributeBody         MessageAttribute = "BODY"
	MessageAttributeBodyPeek     MessageAttribute = "BODY.PEEK"
	MessageAttributeUID          MessageAttribute = "UID"
	MessageAttributeRFC822       MessageAttribute = "RFC822"
	MessageAttributePreview      MessageAttribute = "PREVIEW"
)

type CompoundMessageAttribute struct {
	Attribute MessageAttribute
	Section   string
	Partial   string
}

func NewCompoundMessageAttribute(attribute string) *CompoundMessageAttribute {
	re := regexp.MustCompile(`[\[\]]`)
	fields := re.Split(attribute, -1)
	cma := &CompoundMessageAttribute{
		Attribute: MessageAttribute(fields[0]),
	}

	l := len(fields)
	if l > 1 {
		cma.Section = fields[1]
	}

	if l > 2 {
		cma.Partial = fields[2]
	}

	return cma
}

type FetchMacro string

const (
	FetchMacroAll  FetchMacro = "ALL"
	FetchMacroFast FetchMacro = "FAST"
	FetchMacroFull FetchMacro = "FULL"
)
