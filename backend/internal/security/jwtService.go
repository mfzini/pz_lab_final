package security

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"example/internal/api"

	"github.com/golang-jwt/jwt/v5"
)

const (
	// AccessTokenTTL: vida corta. Es el que viaja en cada request.
	AccessTokenTTL = 15 * time.Minute
	// RefreshTokenTTL: vida larga. Sólo sirve para pedir tokens nuevos.
	RefreshTokenTTL = 7 * 24 * time.Hour

	issuer = "app"
	// Las audiencias separan los tipos de token: un refresh jamás sirve como access.
	accessAudience  = "access"
	refreshAudience = "refresh"
)

type JwtService struct {
}

type Claims struct {
	Role
	jwt.RegisteredClaims
}

func NewJWTService() *JwtService {
	return &JwtService{}
}

func (this *JwtService) SetUserContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := this.ExtractToken(r)
		if token == "" {
			next.ServeHTTP(w, r)
			return
		}

		claims, err := this.ParseAccessToken(token)
		if err != nil {
			w.Header().Set("WWW-Authenticate", `Bearer error="invalid_token"`)
			api.WriteJson(w, api.NewError("Token inválido o vencido.", http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r.WithContext(withClaims(r.Context(), claims)))
	})
}

func (this *JwtService) ExtractToken(r *http.Request) string {
	scheme, token, ok := strings.Cut(r.Header.Get("Authorization"), " ")
	if !ok || !strings.EqualFold(scheme, "Bearer") {
		return ""
	}
	return strings.TrimSpace(token)
}

func (this *JwtService) GenerateAccessToken(id string, role Role) string {
	return this.sign(id, role, AccessTokenTTL, accessAudience, accessSecret())
}

func (this *JwtService) GenerateRefreshToken(id string, role Role) string {
	return this.sign(id, role, RefreshTokenTTL, refreshAudience, refreshSecret())
}

func (this *JwtService) ParseAccessToken(token string) (*Claims, error) {
	return this.parse(token, accessAudience, accessSecret())
}

func (this *JwtService) ParseRefreshToken(token string) (*Claims, error) {
	return this.parse(token, refreshAudience, refreshSecret())
}

func (this *JwtService) parse(token, audience string, secret []byte) (*Claims, error) {
	claims := &Claims{}
	parsed, err := jwt.ParseWithClaims(token, claims,
		func(t *jwt.Token) (any, error) {
			if t.Method != jwt.SigningMethodHS256 {
				return nil, errors.New("método de firma inválido")
			}
			return secret, nil
		},
		jwt.WithIssuer(issuer),
		jwt.WithAudience(audience),
		jwt.WithExpirationRequired(),
	)
	if err != nil {
		return nil, err
	}
	if !parsed.Valid {
		return nil, errors.New("token inválido")
	}
	return claims, nil
}

func (this *JwtService) sign(id string, role Role, ttl time.Duration, audience string, secret []byte) string {
	now := time.Now().UTC()
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		Role: role,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    issuer,
			Audience:  jwt.ClaimStrings{audience},
			Subject:   id,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
	})
	token, err := t.SignedString(secret)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return ""
	}
	return token
}

func accessSecret() []byte {
	if s := os.Getenv("JWT_SECRET"); s != "" {
		return []byte(s)
	}
	return []byte("secret")
}

func refreshSecret() []byte {
	if s := os.Getenv("JWT_REFRESH_SECRET"); s != "" {
		return []byte(s)
	}
	return accessSecret()
}
