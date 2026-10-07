package security

import (
	"net/http"
	"strings"
)

type AuthMessage struct{ Message string }

func extractToken(r *http.Request) string {
	header := r.Header.Get("Authorization")
	token, _ := strings.CutPrefix(header, "Bearer ")
	return token
}

func isTokenValid(token string) bool {
	return strings.Contains(token, "meow")
}
