package security

import (
	"net/http"
	"slices"

	"example/internal/api"
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
			w.Header().Set("WWW-Authenticate", `Bearer`)
			api.WriteJson(w, api.NewError("Se requiere autenticación.", http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		if slices.Contains(roles, claims.Role) {
			next.ServeHTTP(w, r)
			return
		}

		api.WriteJson(w, api.NewError("No tenés permisos para esta operación.", http.StatusForbidden), http.StatusForbidden)
	})
}
