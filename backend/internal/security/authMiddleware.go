package security

import (
	"encoding/json"
	"net/http"
	"strings"
)

type AuthMessage struct{ Message string }

func extractToken(r *http.Request) (string, bool) {
	header := r.Header.Get("Authorization")
	return strings.CutPrefix(header, "Bearer ")
}

func isTokenValid(token string) bool {
	return strings.Contains(token, "meow")
}

func OnlyAdmin(next http.HandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, isPresent := extractTokenFromCookie(r)
		if !isPresent || !isTokenValid(token) {
			w.WriteHeader(403)
			json.NewEncoder(w).Encode(&AuthMessage{Message: "Forbidden"})
			return
		}
		next.ServeHTTP(w, r)
	})
}
