package imap

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
