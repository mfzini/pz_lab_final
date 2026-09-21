package security

import (
	"net/http"
	"strings"
)

func MakeCookie(token string) *http.Cookie {
	return &http.Cookie{Name: "token", Value: token,
		HttpOnly: true,
		Secure:   false,
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
	}
}

func extractTokenFromCookie(r *http.Request) (string, bool) {
	cookie, err := r.Cookie("token")
	if err != nil {
		return "", false
	}
	return strings.CutPrefix(cookie.Value, "Bearer ")
}
