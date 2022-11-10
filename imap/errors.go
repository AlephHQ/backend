package imap

import "errors"

var ErrNotStatusRespCode = errors.New("not a status response code")
var ErrStatusNotOK = errors.New("status not ok")
var ErrFoundSpecialChar = errors.New("found a special char")
var ErrNotSpecialChar = errors.New("found a non-special char")
var ErrUnhandled = errors.New("unhandled response")

// State Errors
var ErrNotAuthenticated = errors.New("not authenticated")
var ErrNotSelected = errors.New("not selected")
