package security

import "context"

// claimsKey es un tipo propio: las claves de contexto de tipo string chocan con
// las de otros paquetes (y go vet lo avisa).
type claimsKey struct{}

func withClaims(ctx context.Context, claims *Claims) context.Context {
	return context.WithValue(ctx, claimsKey{}, claims)
}

func GetClaimsFromContext(ctx context.Context) *Claims {
	if claims, ok := ctx.Value(claimsKey{}).(*Claims); ok {
		return claims
	}
	return nil
}
