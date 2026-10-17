package security

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type jwtService struct {
}

type Claims struct {
	Role
	jwt.RegisteredClaims
}

func NewJWTService() *jwtService {
	return &jwtService{}
}

func (this *jwtService) SetUserContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := this.extractToken(r)
		claims, err := this.getClaims(token)
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}

		rwc := r.WithContext(context.WithValue(r.Context(), "claims", claims))
		next.ServeHTTP(w, rwc)
	})
}

func (this *jwtService) extractToken(r *http.Request) string {
	header := r.Header.Get("Authorization")
	token, _ := strings.CutPrefix(header, "Bearer ")
	return token
}

func (this *jwtService) getClaims(token string) (*Claims, error) {
	t, err := jwt.ParseWithClaims(token, &Claims{}, func(t *jwt.Token) (any, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, errors.New("Wrong signing method!")
		}
		return []byte("secret"), nil
	})
	if err != nil {
		return nil, err
	}
	return t.Claims.(*Claims), nil
}

func (this *jwtService) GenerateToken(id string, role Role) string {

	timestamp := time.Now().UTC()
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		Issuer:    "app",
		Audience:  jwt.ClaimStrings{"user"},
		IssuedAt:  jwt.NewNumericDate(timestamp),
		ExpiresAt: jwt.NewNumericDate(timestamp.AddDate(0, 0, 7)),
		Subject:   id,
		Role:      role,
	})
	token, err := t.SignedString([]byte("secret"))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return ""
	}
	return token
}
