package users

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

type repo struct {
	db *pgx.Conn
}

func NewRepo(connection *pgx.Conn) *repo {
	return &repo{db: connection}
}

func (this *repo) create(u *Model) error {
	query := `insert into users (id, name, surname, email, phone, password, created_at)
		values (@id, @name, @surname, @email, @phone, @password, @created_at)`
	timestamp := time.Now()
	_, err := this.db.Exec(context.Background(), query,
		&pgx.NamedArgs{
			"id":         u.Id,
			"name":       u.Name,
			"surname":    u.Surname,
			"email":      u.Email,
			"phone":      u.Phone,
			"password":   u.Password,
			"created_at": timestamp,
		})
	if err == nil {
		u.CreatedAt = &timestamp
	}
	return err
}

func (this *repo) updateField(ctx context.Context, field string, value any, id string) error {
	query := fmt.Sprint("update users set %s = $1, updated_at = $2 where id = $3", field)
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
	query := "update users set deleted_at=$1 where id=$2"
	timestamp := time.Now()
	_, err := this.db.Exec(context.Background(), query, timestamp, u.Id)
	if err == nil {
		u.DeletedAt = &timestamp
	}
	return err
}

func (this *repo) FindById(id string) (*Model, error) {
	query := "select * from users where id = $1"
	r, err := this.db.Query(context.Background(), query, id)
	if err != nil {
		return nil, err
	}
	return pgx.CollectOneRow(r, pgx.RowToAddrOfStructByName[Model])
}

func (this *repo) FindByEmail(email string) (*Model, error) {
	query := "select * from users where email = $1"
	r, err := this.db.Query(context.Background(), query, email)
	if err != nil {
		return nil, err
	}
	return pgx.CollectOneRow(r, pgx.RowToAddrOfStructByName[Model])
}
