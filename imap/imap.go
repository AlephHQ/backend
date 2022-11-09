package imap

import "errors"

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

type MessageStatusResponseCode string

const (
	MessageStatusResponseCodeFetch MessageStatusResponseCode = "FETCH"
)

var ErrNotStatusRespCode = errors.New("not a status response code")
var ErrStatusNotOK = errors.New("status not ok")
var ErrFoundSpecialChar = errors.New("found a special char")
var ErrNotSpecialChar = errors.New("found a non-special char")
var ErrUnhandled = errors.New("unhandled response")

type MessageAttribute string

const (
	MessageAttributeFlags        MessageAttribute = "FLAGS"
	MessageAttributeInternalDate MessageAttribute = "INTERNALDATE"
	MessageAttributeRFC822Size   MessageAttribute = "RFC822.SIZE"
	MessageAttributeEnvelope     MessageAttribute = "ENVELOPE"
	MessageAttributeBody         MessageAttribute = "BODY"
)

type FetchMacro string

const (
	FetchMacroAll  FetchMacro = "ALL"
	FetchMacroFast FetchMacro = "FAST"
	FetchMacroFull FetchMacro = "FULL"
)
