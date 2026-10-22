package users

import (
	"encoding/json"
	"example/internal/api"
	"example/internal/domain/common"
	"example/internal/security"
	"fmt"
	"net/http"
	"os"
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

type RefreshRequest struct {
	RefreshToken string `json:"refreshToken"`
}

type LoginResponse struct {
	Token        string `json:"token"`
	RefreshToken string `json:"refreshToken"`
	User         *Model `json:"user"`
}

func (this *handler) Setup(mux *http.ServeMux) {

	mux.HandleFunc("POST /login", this.login)
	mux.HandleFunc("POST /refresh", this.refresh)

	mux.HandleFunc("POST /user", this.create)
	mux.HandleFunc("PATCH /user/{id}", security.AllowClients(this.update))
	mux.HandleFunc("PATCH /user/{id}/password", security.AllowClients(this.updatePassword))
	mux.HandleFunc("PATCH /user/{id}/email", security.AllowClients(this.updateEmail))
	mux.HandleFunc("PATCH /user/{id}/phone", security.AllowClients(this.updatePhone))
	mux.HandleFunc("PATCH /user/{id}/avatar", security.AllowClients(this.updateAvatar))
	mux.HandleFunc("POST /user/{id}/avatar", security.AllowClients(this.uploadAvatar))
	mux.HandleFunc("PATCH /user/{id}/role", security.OnlyAdmin(this.updateRole))
	mux.HandleFunc("GET /user/{id}", security.AllowClients(this.getByID))
	// Listado de usuarios: sólo admin (panel de roles).
	mux.HandleFunc("GET /usuarios", security.OnlyAdmin(this.list))

	mux.HandleFunc("POST /init", func(w http.ResponseWriter, r *http.Request) {
		if this.Service.AdminExists() {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		data := &LoginRequest{}
		if err := json.NewDecoder(r.Body).Decode(data); err != nil {
			api.WriteJson(w, api.NewError("Body inválido.", http.StatusBadRequest), http.StatusBadRequest)
			return
		}
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
		// Create normaliza el rol a "client": acá lo dejamos como quedó en la DB,
		// así los tokens salen con el rol real.
		admin.Role = security.Admin
		this.writeTokens(w, admin, http.StatusCreated)
	})

}

func (this *handler) login(w http.ResponseWriter, r *http.Request) {
	req := &LoginRequest{}
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		api.WriteJson(w, api.NewError("Body inválido.", http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	user, err := this.Service.FindByEmail(req.Email)
	if err != nil {
		api.WriteJson(w, api.NewGenericError(err), http.StatusInternalServerError)
		return
	}
	if user == nil || user.ComparePassword(req.Password) != nil {
		api.WriteJson(w, api.NewError("Credenciales inválidas.", http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	this.writeTokens(w, user, http.StatusOK)
}

func (this *handler) refresh(w http.ResponseWriter, r *http.Request) {
	req := &RefreshRequest{}
	decodeErr := json.NewDecoder(r.Body).Decode(req)

	token := req.RefreshToken
	if token == "" {
		token = this.jwtService.ExtractToken(r)
	}
	if token == "" {
		if decodeErr != nil {
			api.WriteJson(w, api.NewError("Body inválido.", http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		api.WriteJson(w, api.NewError("Falta el refresh token.", http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	claims, err := this.jwtService.ParseRefreshToken(token)
	if err != nil {
		w.Header().Set("WWW-Authenticate", `Bearer error="invalid_token"`)
		api.WriteJson(w, api.NewError("Refresh token inválido o vencido.", http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	user, err := this.Service.FindById(claims.Subject)
	if err != nil {
		api.WriteJson(w, api.NewGenericError(err), http.StatusInternalServerError)
		return
	}
	if user == nil {
		api.WriteJson(w, api.NewError("Sesión inválida.", http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	this.writeTokens(w, user, http.StatusOK)
}

func (this *handler) writeTokens(w http.ResponseWriter, user *Model, statusCode int) {
	access := this.jwtService.GenerateAccessToken(user.Id, user.Role)
	refresh := this.jwtService.GenerateRefreshToken(user.Id, user.Role)
	w.Header().Set("Authorization", "Bearer "+access)
	api.WriteJson(w, &LoginResponse{
		Token:        access,
		RefreshToken: refresh,
		User:         user,
	}, statusCode)
}

func (this *handler) create(w http.ResponseWriter, r *http.Request) {
	user := &Model{}
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := this.Service.Create(user); err != nil {
		api.WriteJson(w, err, err.StatusCode)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (this *handler) update(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if !guardOwner(w, r, id) {
		return
	}

	patch := &PatchRequest{}
	if !decodeBody(w, r, patch) {
		return
	}

	user, err := this.Service.Update(r.Context(), id, patch)
	if err != nil {
		api.WriteJson(w, err, err.StatusCode)
		return
	}
	api.WriteJson(w, user, http.StatusOK)
}

// PATCH /user/{id}/password — body: {"password": "..."}
func (this *handler) updatePassword(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if !guardOwner(w, r, id) {
		return
	}
	body := &struct {
		Password string `json:"password"`
	}{}
	if !decodeBody(w, r, body) {
		return
	}
	if err := this.Service.UpdatePassword(r.Context(), id, body.Password); err != nil {
		api.WriteJson(w, err, err.StatusCode)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// PATCH /user/{id}/email — body: {"email": "..."}
func (this *handler) updateEmail(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if !guardOwner(w, r, id) {
		return
	}
	body := &struct {
		Email string `json:"email"`
	}{}
	if !decodeBody(w, r, body) {
		return
	}
	if err := this.Service.UpdateEmail(r.Context(), id, body.Email); err != nil {
		api.WriteJson(w, err, err.StatusCode)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// PATCH /user/{id}/phone — body: {"phone": "..."}
func (this *handler) updatePhone(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if !guardOwner(w, r, id) {
		return
	}
	body := &struct {
		Phone string `json:"phone"`
	}{}
	if !decodeBody(w, r, body) {
		return
	}
	if err := this.Service.UpdatePhone(r.Context(), id, body.Phone); err != nil {
		api.WriteJson(w, err, err.StatusCode)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// PATCH /user/{id}/avatar — body: {"avatar": "/uploads/x.jpg"} (null o "" = sin avatar)
func (this *handler) updateAvatar(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if !guardOwner(w, r, id) {
		return
	}
	body := &struct {
		Avatar *string `json:"avatar"`
	}{}
	if !decodeBody(w, r, body) {
		return
	}
	if err := this.Service.UpdateAvatar(r.Context(), id, body.Avatar); err != nil {
		api.WriteJson(w, err, err.StatusCode)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// POST /user/{id}/avatar — multipart con "picture": sube la imagen, la guarda
// en /uploads y la deja como avatar del usuario (dueño o admin).
func (this *handler) uploadAvatar(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if !guardOwner(w, r, id) {
		return
	}
	file, _, err := r.FormFile("picture")
	if err != nil {
		api.WriteJson(w, api.NewError("Falta la foto.", http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	defer file.Close()

	avatar, apiErr := this.Service.UploadAvatar(r.Context(), id, file)
	if apiErr != nil {
		api.WriteJson(w, apiErr, apiErr.StatusCode)
		return
	}
	api.WriteJson(w, &struct {
		Avatar *string `json:"avatar"`
	}{avatar}, http.StatusOK)
}

// PATCH /user/{id}/role — body: {"role": "admin"}. Sólo admin (en la ruta).
func (this *handler) updateRole(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	body := &struct {
		Role security.Role `json:"role"`
	}{}
	if !decodeBody(w, r, body) {
		return
	}
	if err := this.Service.UpdateRole(r.Context(), id, body.Role); err != nil {
		api.WriteJson(w, err, err.StatusCode)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// GET /usuarios?page=&size= — sólo admin: listado paginado con el rol de
// cada usuario, para poder asignarlo desde el panel.
func (this *handler) list(w http.ResponseWriter, r *http.Request) {
	page, size := common.PageParams(r.URL.Query())

	users, pages, apiErr := this.Service.List(r.Context(), page, size)
	if apiErr != nil {
		api.WriteJson(w, apiErr, apiErr.StatusCode)
		return
	}
	api.WriteJson(w, &common.PageResponse[*Model]{
		Data:        users,
		CurrentPage: page,
		TotalPages:  pages,
	}, http.StatusOK)
}

func (this *handler) getByID(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	if !guardOwner(w, r, id) {
		return
	}

	p, err := this.Service.FindById(id)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if p == nil {
		api.WriteJson(w, api.NewError("Usuario no encontrado.", http.StatusNotFound), http.StatusNotFound)
		return
	}
	api.WriteJson(w, p, http.StatusOK)
}

// canAccess: dueño del recurso o admin. El middleware ya garantiza que hay
// claims, pero el chequeo no cuesta nada.
func canAccess(r *http.Request, id string) bool {
	claims := security.GetClaimsFromContext(r.Context())
	return claims != nil && (claims.Subject == id || claims.Role == security.Admin)
}

// guardOwner: si no es dueño ni admin escribe el 403 y devuelve false.
func guardOwner(w http.ResponseWriter, r *http.Request, id string) bool {
	if canAccess(r, id) {
		return true
	}
	api.WriteJson(w, api.NewError("No tenés permisos para esta operación.", http.StatusForbidden), http.StatusForbidden)
	return false
}

// decodeBody devuelve false (y manda el 400) si el body no se pudo leer.
func decodeBody(w http.ResponseWriter, r *http.Request, dst any) bool {
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		api.WriteJson(w, api.NewError("Body inválido.", http.StatusBadRequest), http.StatusBadRequest)
		return false
	}
	return true
}
