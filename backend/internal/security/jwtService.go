package security

import (
	"context"
	"errors"
	"example/internal/domain/users"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type jwtService struct {
	userService *users.Service
}

type Claims struct {
	Role string
	jwt.RegisteredClaims
}

func NewJWTService(usersService *users.Service) *jwtService {
	return &jwtService{usersService}
}

func (this *jwtService) SetUserContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := this.extractToken(r)
		claims, err := this.getClaims(token)
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}

		userId, err := claims.GetSubject()
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}
		user, err := this.userService.FindById(userId)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		rwc := r.WithContext(context.WithValue(r.Context(), "user", user))
		next.ServeHTTP(w, rwc)
	})
}

func (this *jwtService) extractToken(r *http.Request) string {
	header := r.Header.Get("Authorization")
	token, _ := strings.CutPrefix(header, "Bearer ")
	return token
}

func (this *jwtService) getClaims(token string) (jwt.Claims, error) {
	t, err := jwt.ParseWithClaims(token, &Claims{}, func(t *jwt.Token) (any, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, errors.New("Wrong signing method!")
		}
		return []byte("secret"), nil
	})
	if err != nil {
		return nil, err
	}
	return t.Claims, nil
}

func (this *jwtService) GenerateToken(user *users.Model) string {

	timestamp := time.Now().UTC()
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		Issuer:    "app",
		Audience:  jwt.ClaimStrings{"user"},
		IssuedAt:  jwt.NewNumericDate(timestamp),
		ExpiresAt: jwt.NewNumericDate(timestamp.AddDate(0, 0, 7)),
		Subject:   user.Id,
		Role:      user.Role,
	})
	token, err := t.SignedString([]byte("secret"))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return ""
	}
	return token
}
