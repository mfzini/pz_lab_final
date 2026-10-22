package users

import (
	"context"
	"fmt"
	"time"

	"example/internal/domain/common"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Columnas de lectura del listado: sin password (además el Model lo manda
// con json:"-", pero no hace falta ni traerlo de la DB). Se usa Lax porque
// el struct tiene el campo Password que esta query no trae.
const columns = `id, name, surname, email, phone, avatar, role, created_at, updated_at`

type repo struct {
	db *pgxpool.Pool
}

func NewRepo(connection *pgxpool.Pool) *repo {
	return &repo{db: connection}
}

func (this *repo) create(u *Model) error {
	query := `insert into users (id, name, surname, email, phone, password, role, created_at)
		values (@id, @name, @surname, @email, @phone, @password, @role, @created_at)`
	timestamp := time.Now()
	_, err := this.db.Exec(context.Background(), query,
		&pgx.NamedArgs{
			"id":         u.Id,
			"name":       u.Name,
			"surname":    u.Surname,
			"email":      u.Email,
			"phone":      u.Phone,
			"password":   u.Password,
			"role":       string(u.Role),
			"created_at": timestamp,
		})
	if err == nil {
		u.CreatedAt = &timestamp
	}
	return err
}

// updateField actualiza una columna suelta. Como el nombre va pegado a la query,
// sólo se aceptan los campos de este listado: nada de concatenar lo que llegue.
func (this *repo) updateField(ctx context.Context, field string, value any, id string) error {
	switch field {
	case "role", "phone", "avatar", "password", "email":
	default:
		return fmt.Errorf("campo no editable: %q", field)
	}
	query := fmt.Sprintf("update users set %s = $1, updated_at = $2 where id = $3", field)
	_, err := this.db.Exec(ctx, query, value, time.Now(), id)
	return err
}

func (this *repo) update(u *Model) error {
	query := `update users set
	name = @name,
	surname = @surname,
	updated_at = @now
	where id = @id`
	timestamp := time.Now()
	_, err := this.db.Exec(context.Background(), query, &pgx.NamedArgs{
		"name":    u.Name,
		"surname": u.Surname,
		"now":     timestamp,
		"id":      u.Id,
	})
	if err == nil {
		u.UpdatedAt = &timestamp
	}
	return err
}

func (this *repo) delete(u *Model) error {
	query := "update users set deleted_at = $1 where id = $2"
	timestamp := time.Now()
	_, err := this.db.Exec(context.Background(), query, timestamp, u.Id)
	if err == nil {
		u.DeletedAt = &timestamp
	}
	return err
}

func (this *repo) FindById(id string) (*Model, error) {
	query := "select * from users where id = $1 and deleted_at is null"
	r, err := this.db.Query(context.Background(), query, id)
	if err != nil {
		return nil, err
	}
	return pgx.CollectOneRow(r, pgx.RowToAddrOfStructByName[Model])
}

func (this *repo) FindByEmail(email string) (*Model, error) {
	query := "select * from users where email = $1 and deleted_at is null"
	r, err := this.db.Query(context.Background(), query, email)
	if err != nil {
		return nil, err
	}
	return pgx.CollectOneRow(r, pgx.RowToAddrOfStructByName[Model])
}

// list es el listado para el panel del admin (paginado, sin password).
func (this *repo) list(ctx context.Context, page, size int) ([]*Model, int, error) {
	query := fmt.Sprintf(`select %s from users where deleted_at is null
		order by name, surname limit $1 offset $2`, columns)
	rows, err := this.db.Query(ctx, query, size, (page-1)*size)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	users, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByNameLax[Model])
	if err != nil {
		return nil, 0, err
	}

	totalRows := 0
	if err := this.db.QueryRow(ctx,
		`select count(*) from users where deleted_at is null`).Scan(&totalRows); err != nil {
		return nil, 0, err
	}
	return users, common.TotalPagesFor(totalRows, size), nil
}
