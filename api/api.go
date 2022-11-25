package api

type ContextKeyName string

const (
	ContextKeyNameParams ContextKeyName = "params"
	ContextKeyNameUserID ContextKeyName = "user_id"
)

type CookieName string

const (
	CookieNameSession CookieName = "ssid"
)
