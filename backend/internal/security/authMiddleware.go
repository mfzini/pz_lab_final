package security

import (
	"example/internal/domain/users"
	"net/http"
)

func OnlyAdmin(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		uctx := r.Context().Value("user")
		if uctx == nil {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		user := uctx.(*users.Model)
		if user.Role != "admin" {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}
