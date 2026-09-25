package repositories

import (
	"context"
	"example/internal/models"
	"time"

	"github.com/jackc/pgx/v5"
)

type PeopleRepo struct {
	Db *pgx.Conn
}

func NewPeopleRepo(connection *pgx.Conn) *PeopleRepo {
	return &PeopleRepo{Db: connection}
}

func (self *PeopleRepo) Create(p *models.Person) error {
	query := `insert into people (id, name, surname, email, phone, created_at)
		values (@id, @name, @surname, @email, @phone, @created_at)`
	timestamp := time.Now()
	_, err := self.Db.Exec(context.Background(), query,
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

func (self *PeopleRepo) Update(p *models.Person) error {
	query := `update people set
	name = @name,
	surname = @surname,
	email = @email,
	phone = @phone,
	updated_at = @now
	where id = @id`
	timestamp := time.Now()
	_, err := self.Db.Exec(context.Background(), query, &pgx.NamedArgs{
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

func (self *PeopleRepo) Delete(p *models.Person) error {
	query := "update people set deleted_at=$1 where id=$2"
	timestamp := time.Now()
	_, err := self.Db.Exec(context.Background(), query, timestamp, p.Id)
	if err == nil {
		p.DeletedAt = &timestamp
	}
	return err
}

func (self *PeopleRepo) FindById(id string) (*models.Person, error) {
	query := "select * from people where id = $1"
	r, err := self.Db.Query(context.Background(), query, id)
	if err != nil {
		return nil, err
	}
	return pgx.CollectOneRow(r, pgx.RowToAddrOfStructByName[models.Person])
}

func (self *PeopleRepo) FindByEmail(email string) (*models.Person, error) {
	query := "select * from people where email = $1"
	r, err := self.Db.Query(context.Background(), query, email)
	if err != nil {
		return nil, err
	}
	return pgx.CollectOneRow(r, pgx.RowToAddrOfStructByName[models.Person])
}
