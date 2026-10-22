package users

import (
	"context"
	"errors"

	"example/internal/api"
	"example/internal/domain/images"
	"example/internal/security"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/mail"

	"github.com/jackc/pgx/v5"
	"uuid"
)

type Service struct {
	repo          *repo
	imagesService *images.Service
}

func NewService(repo *repo, imagesService *images.Service) *Service {
	return &Service{repo, imagesService}
}

func (this *Service) Create(user *Model) *api.Error {
	if user.Id == "" {
		user.Id = uuid.NewV7().String()
	}
	// El rol lo pone el servidor: lo que venga en el body se descarta.
	user.Role = security.User
	if user.Name == "" {
		return api.NewError("El nombre no puede estar vacio.", http.StatusBadRequest)
	}
	if user.Surname == "" {
		return api.NewError("El apellido no puede estar vacio.", http.StatusBadRequest)
	}
	if user.Email == "" {
		return api.NewError("El email no puede estar vacio.", http.StatusBadRequest)
	}
	if _, err := mail.ParseAddress(user.Email); err != nil {
		return api.NewError("Formato de email no válido.", http.StatusBadRequest)
	}

	if exists, _ := this.repo.FindByEmail(user.Email); exists != nil {
		return api.NewError("Email no válido.", http.StatusNotAcceptable)
	}

	if err := user.hashPassword(); err != nil {
		return api.NewError("No se pudo procesar la contraseña.", http.StatusInternalServerError)
	}
	if err := this.repo.create(user); err != nil {
		return api.NewGenericError(err)
	}

	return nil
}

// Update aplica un patch de nombre y apellido sobre la fila existente: los
// campos que no vienen en el body quedan como estaban. Email, contraseña,
// avatar y rol van por sus propios métodos. Sólo el dueño o un admin.
func (this *Service) Update(ctx context.Context, id string, patch *PatchRequest) (*Model, *api.Error) {
	user, err := this.repo.FindById(id)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, api.NewError("Usuario no encontrado.", http.StatusNotFound)
	}
	if err != nil {
		return nil, api.NewGenericError(err)
	}

	claims := security.GetClaimsFromContext(ctx)
	if claims == nil {
		return nil, api.NewError("Se requiere autenticación.", http.StatusUnauthorized)
	}
	if claims.Subject != id && claims.Role != security.Admin {
		return nil, api.NewError("No podés modificar este usuario.", http.StatusForbidden)
	}

	if patch.Name != nil {
		if *patch.Name == "" {
			return nil, api.NewError("El nombre no puede estar vacio.", http.StatusBadRequest)
		}
		user.Name = *patch.Name
	}
	if patch.Surname != nil {
		if *patch.Surname == "" {
			return nil, api.NewError("El apellido no puede estar vacio.", http.StatusBadRequest)
		}
		user.Surname = *patch.Surname
	}

	if err := this.repo.update(user); err != nil {
		return nil, api.NewGenericError(err)
	}
	return user, nil
}

func (this *Service) FindById(id string) (*Model, error) {
	user, err := this.repo.FindById(id)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil // no existe ≠ error de infraestructura
	}
	return user, err
}
func (this *Service) FindByEmail(email string) (*Model, error) {
	user, err := this.repo.FindByEmail(email)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return user, err
}

func (this *Service) AdminExists() bool {
	count := 0
	r := this.repo.db.QueryRow(context.Background(),
		"select count(*) from users where role = 'admin' and deleted_at is null;")
	if err := r.Scan(&count); err != nil {
		return false
	}
	return count > 0
}

func (this *Service) UpdateRole(ctx context.Context, id string, role security.Role) *api.Error {
	if role != security.Admin && role != security.User {
		return api.NewError("Rol no válido.", http.StatusBadRequest)
	}
	// Un admin no puede quitarse su propio rol: se quedaría sin poder entrar
	// a nada, y desde la UI ya no tendría cómo recuperarlo.
	if claims := security.GetClaimsFromContext(ctx); claims != nil && claims.Subject == id && role != security.Admin {
		return api.NewError("No podés quitar tu propio rol de administrador.", http.StatusBadRequest)
	}
	if err := this.repo.updateField(ctx, "role", string(role), id); err != nil {
		return api.NewGenericError(err)
	}
	return nil
}

// List es el listado paginado de usuarios para el panel del admin.
func (this *Service) List(ctx context.Context, page, size int) ([]*Model, int, *api.Error) {
	users, pages, err := this.repo.list(ctx, page, size)
	if err != nil {
		return nil, 0, api.NewGenericError(err)
	}
	return users, pages, nil
}

// UpdatePhone cambia el teléfono con la misma pauta que el email.
func (this *Service) UpdatePhone(ctx context.Context, id, phone string) *api.Error {
	if phone == "" {
		return api.NewError("El teléfono no puede estar vacio.", http.StatusBadRequest)
	}
	if len(phone) > 20 { // la columna es varchar(20): mejor un 400 que un 500
		return api.NewError("El teléfono no puede superar los 20 caracteres.", http.StatusBadRequest)
	}
	if err := this.repo.updateField(ctx, "phone", phone, id); err != nil {
		return api.NewGenericError(err)
	}
	return nil
}

// UpdateEmail cambia el email con las mismas validaciones del registro.
func (this *Service) UpdateEmail(ctx context.Context, id, email string) *api.Error {
	if email == "" {
		return api.NewError("El email no puede estar vacio.", http.StatusBadRequest)
	}
	if _, err := mail.ParseAddress(email); err != nil {
		return api.NewError("Formato de email no válido.", http.StatusBadRequest)
	}
	if exists, _ := this.repo.FindByEmail(email); exists != nil && exists.Id != id {
		return api.NewError("Email no válido.", http.StatusNotAcceptable)
	}
	if err := this.repo.updateField(ctx, "email", email, id); err != nil {
		return api.NewGenericError(err)
	}
	return nil
}

// UpdatePassword cambia la contraseña por su cuenta, hasheada. Se valida acá
// que no venga vacía; que sea la del propio usuario lo garantiza quien llame.
func (this *Service) UpdatePassword(ctx context.Context, id, plain string) *api.Error {
	if plain == "" {
		return api.NewError("La contraseña no puede estar vacia.", http.StatusBadRequest)
	}
	m := &Model{Password: plain}
	if err := m.hashPassword(); err != nil {
		return api.NewError("No se pudo procesar la contraseña.", http.StatusInternalServerError)
	}
	if err := this.repo.updateField(ctx, "password", m.Password, id); err != nil {
		return api.NewGenericError(err)
	}
	return nil
}

// UpdateAvatar setea la imagen de perfil. Nil (o "") la deja sin avatar.
func (this *Service) UpdateAvatar(ctx context.Context, id string, avatar *string) *api.Error {
	var value any // string, o nil para dejarlo sin avatar
	if avatar != nil && *avatar != "" {
		value = *avatar
	}
	if err := this.repo.updateField(ctx, "avatar", value, id); err != nil {
		return api.NewGenericError(err)
	}
	return nil
}

// UploadAvatar sube la foto de perfil (multipart) y la deja como avatar.
// Save ya valida que el archivo sea imagen; sólo el dueño o un admin llegan
// acá (guardOwner en el handler).
func (this *Service) UploadAvatar(ctx context.Context, id string, img multipart.File) (*string, *api.Error) {
	hash, err := this.imagesService.Save(img)
	if err != nil {
		// "Solo se admiten imágenes" viene como *api.Error con su 400: no
		// lo aplastamos a un 500 genérico.
		var apiErr *api.Error
		if errors.As(err, &apiErr) {
			return nil, apiErr
		}
		return nil, api.NewGenericError(err)
	}
	path := fmt.Sprintf("/uploads/%s", hash)
	if err := this.repo.updateField(ctx, "avatar", path, id); err != nil {
		return nil, api.NewGenericError(err)
	}
	return &path, nil
}
