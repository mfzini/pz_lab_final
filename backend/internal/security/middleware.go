package security

import (
	"net/http"
)

func OnlyAdmin(next http.HandlerFunc) http.HandlerFunc {
	return Allow([]Role{Admin}, next)
}

func AllowClients(next http.HandlerFunc) http.HandlerFunc {
	return Allow([]Role{Admin, User}, next)
}

func Allow(roles []Role, next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims := GetClaimsFromContext(r.Context())
		if claims == nil {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		allow := false
		for _, role := range roles {
			if claims.Role == Role(role) {
				allow = true
				break
			}
		}
		if !allow {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}
