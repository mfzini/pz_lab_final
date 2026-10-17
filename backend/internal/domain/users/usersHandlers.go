package users

import (
	"encoding/json"
	"example/internal/api"
	"example/internal/security"
	"fmt"
	"net/http"
	"os"

	"golang.org/x/crypto/bcrypt"
)

type Response struct {
	Location string
}

type handler struct {
	*Service
	jwtService *security.JwtService
}

func NewHandler(Service *Service, jwtService *security.JwtService) *handler {
	return &handler{Service, jwtService}
}

type LoginRequest struct {
	Email,
	Password string
}

func (this *handler) Setup(mux *http.ServeMux) {
	mux.HandleFunc("POST /user", this.create)

	mux.HandleFunc("PATCH /user/{id}", this.update)
	mux.HandleFunc("GET /user/{id}", this.getByID)

	mux.HandleFunc("POST /init", func(w http.ResponseWriter, r *http.Request) {
		if this.Service.AdminExists() {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		data := &LoginRequest{}
		json.NewDecoder(r.Body).Decode(data)
		admin := &Model{
			Name:     "Cosme",
			Surname:  "Fulanito",
			Email:    data.Email,
			Password: data.Password,
			Role:     "admin",
		}
		if err := this.Service.Create(admin); err != nil {
			w.WriteHeader(err.StatusCode)
			return
		}
		if err := this.Service.UpdateRole(r.Context(), admin.Id, security.Admin); err != nil {
			w.WriteHeader(err.StatusCode)
			return
		}
	})

	mux.HandleFunc("POST /login", func(w http.ResponseWriter, r *http.Request) {
		loginRequest := &LoginRequest{}
		json.NewDecoder(r.Body).Decode(loginRequest)
		user, err := this.Service.FindByEmail(loginRequest.Email)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(loginRequest.Password)); err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		token := this.jwtService.GenerateToken(user.Id, security.Role(user.Role))
		w.Header().Add("Authorization", "Bearer "+token)
		w.WriteHeader(http.StatusOK)
	})
}

func (this *handler) create(w http.ResponseWriter, r *http.Request) {
	user := &Model{}
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := this.Service.Create(user); err != nil {
		w.WriteHeader(err.StatusCode)
		fmt.Fprintln(os.Stderr, err)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (this *handler) update(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	data := &Model{}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	data.Id = id
	if err := this.Service.Update(r.Context(), id, data); err != nil {
		api.WriteJson(w, err, err.StatusCode)
	}
}

func (this *handler) getByID(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	p, err := this.Service.repo.FindById(id)
	if err != nil {
		w.WriteHeader(500)
		fmt.Fprintf(os.Stderr, "%+v", err)
		return
	}
	if p == nil {
		w.WriteHeader(404)
		return
	}
	w.WriteHeader(200)
	json.NewEncoder(w).Encode(p)
}
