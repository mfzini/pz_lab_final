package users

import (
	"example/internal/domain/common"
	"example/internal/security"

	"golang.org/x/crypto/bcrypt"
)

type Model struct {
	Id,
	Name,
	Surname,
	Email string
	// La columna es anulable: sin teléfono se lee null, no "".
	Phone    *string
	Role     security.Role
	Avatar   *string
	Password string `json:"-"`
	common.TimeStamps
}

// PatchRequest es lo que acepta PATCH /user/{id}: sólo nombre y apellido.
// Email, contraseña, avatar y rol tienen su propio endpoint y método en el
// service; ausentes (o null) dejan el valor actual como está.
type PatchRequest struct {
	Name    *string `json:"name"`
	Surname *string `json:"surname"`
}

func (this *Model) hashPassword() error {
	hash, err := bcrypt.GenerateFromPassword([]byte(this.Password), 10)
	if err != nil {
		return err
	}
	this.Password = string(hash)
	return nil
}

func (this *Model) ComparePassword(password string) error {
	return bcrypt.CompareHashAndPassword([]byte(this.Password), []byte(password))
}
