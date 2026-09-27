package people

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
)

type repo struct {
	db *pgx.Conn
}

func NewRepo(connection *pgx.Conn) *repo {
	return &repo{db: connection}
}

func (this *repo) create(p *Model) error {
	query := `insert into people (id, name, surname, email, phone, created_at)
		values (@id, @name, @surname, @email, @phone, @created_at)`
	timestamp := time.Now()
	_, err := this.db.Exec(context.Background(), query,
		&pgx.NamedArgs{
			"id":         p.Id,
			"name":       p.Name,
			"surname":    p.Surname,
			"email":      p.Email,
			"phone":      p.Phone,
			"created_at": timestamp,
		})
	if err == nil {
		p.CreatedAt = &timestamp
	}
	return err
}

func (this *repo) update(p *Model) error {
	query := `update people set
	name = @name,
	surname = @surname,
	email = @email,
	phone = @phone,
	updated_at = @now
	where id = @id`
	timestamp := time.Now()
	_, err := this.db.Exec(context.Background(), query, &pgx.NamedArgs{
		"name":    p.Name,
		"surname": p.Surname,
		"email":   p.Email,
		"phone":   p.Phone,
		"now":     timestamp,
		"id":      p.Id,
	})
	if err == nil {
		p.UpdatedAt = &timestamp
	}
	return err
}

func (this *repo) delete(p *Model) error {
	query := "update people set deleted_at=$1 where id=$2"
	timestamp := time.Now()
	_, err := this.db.Exec(context.Background(), query, timestamp, p.Id)
	if err == nil {
		p.DeletedAt = &timestamp
	}
	return err
}

func (this *repo) FindById(id string) (*Model, error) {
	query := "select * from people where id = $1"
	r, err := this.db.Query(context.Background(), query, id)
	if err != nil {
		return nil, err
	}
	return pgx.CollectOneRow(r, pgx.RowToAddrOfStructByName[Model])
}

func (this *repo) FindByEmail(email string) (*Model, error) {
	query := "select * from people where email = $1"
	r, err := this.db.Query(context.Background(), query, email)
	if err != nil {
		return nil, err
	}
	return pgx.CollectOneRow(r, pgx.RowToAddrOfStructByName[Model])
}
