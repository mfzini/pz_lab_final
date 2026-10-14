package security

import (
	"encoding/json"
	"example/internal/domain/users"
	"fmt"
	"net/http"

	"golang.org/x/crypto/bcrypt"
)

type loginHandler struct {
	usersService *users.Service
	jwtService   *jwtService
}

func NewAuthHandler(usersService *users.Service, jwtService *jwtService) *loginHandler {
	return &loginHandler{usersService, jwtService}
}

type LoginRequest struct {
	Email,
	Password string
}

func (this *loginHandler) Setup(mux *http.ServeMux) {

	mux.HandleFunc("POST /init", func(w http.ResponseWriter, r *http.Request) {
		if this.usersService.AdminExists() {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		data := &LoginRequest{}
		json.NewDecoder(r.Body).Decode(data)
		admin := &users.Model{
			Name:     "Cosme",
			Surname:  "Fulanito",
			Email:    data.Email,
			Password: data.Password,
			Role:     "admin",
		}
		this.usersService.Create(admin)
		this.usersService.Update(admin)
	})

	mux.HandleFunc("POST /login", func(w http.ResponseWriter, r *http.Request) {
		loginRequest := &LoginRequest{}
		json.NewDecoder(r.Body).Decode(loginRequest)
		user, err := this.usersService.FindByEmail(loginRequest.Email)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Println(err)
			return
		}
		if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(loginRequest.Password)); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Println(err)
			return
		}
		token := this.jwtService.GenerateToken(user)
		w.Header().Add("Authorization", "Bearer: "+token)
		w.WriteHeader(http.StatusOK)
	})
}
