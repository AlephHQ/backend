package api

import (
	"aleph/backend/env"
	"aleph/backend/utils"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

const (
	CookieNameAuthTok = "_nar"
	CookieNameCsrf    = "uto_"
)

type JWTClaims struct {
	CsrfToken string
	jwt.RegisteredClaims
}

func NewJWT(claims JWTClaims) (ss string, err error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	ss, err = token.SignedString(env.SigningKey())
	return
}

func SetAuthCookies(w http.ResponseWriter, u *User) {
	csrf := utils.RandStr(20)
	tok, err := NewJWT(JWTClaims{
		CsrfToken: csrf,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().AddDate(0, 0, 1)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "aleph/api/v1.0",
			Subject:   u.ID.Hex(),
		},
	})
	if err != nil {
		panic("error generating new jwt token:" + err.Error())
	}

	// setting the jwt cookie
	http.SetCookie(
		w,
		&http.Cookie{
			Name:     CookieNameAuthTok,
			Value:    tok,
			MaxAge:   24 * 36000,
			Path:     "/",
			HttpOnly: true,
			Secure:   true,
		},
	)

	// setting the csrf token cookie
	http.SetCookie(
		w,
		&http.Cookie{
			Name:   CookieNameCsrf,
			Value:  csrf,
			MaxAge: 24 * 3600,
			Path:   "/",
			Secure: true,
		},
	)
}

func GetAuthToken(r *http.Request) (t *jwt.Token, err error) {
	var c *http.Cookie
	c, err = r.Cookie(CookieNameAuthTok)
	if err != nil {
		return
	}

	t, err = jwt.ParseWithClaims(
		c.Value,
		&JWTClaims{},
		func(token *jwt.Token) (interface{}, error) {
			return env.SigningKey(), nil
		},
		jwt.WithValidMethods([]string{jwt.SigningMethodHS512.Name}),
	)
	return
}
