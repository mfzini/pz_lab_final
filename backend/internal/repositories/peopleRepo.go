package repositories

import (
	"database/sql"
	"example/internal/db"
	"example/internal/models"
	"time"
)

type PeopleRepo struct{}

func (repo *PeopleRepo) Save(p *models.Person) (int64, error) {
	query := "insert into people (id, name, surname, email, phone) values (?, ?, ?, ?, ?)"
	stmt, _ := db.PrepareStatement(query)
	r, err := stmt.Exec(p.Id.String(), p.Name, p.Surname, p.Email, p.Phone)
	if err != nil {
		return 0, err
	}
	return r.RowsAffected()
}

func (repo *PeopleRepo) FindById(id string) (*models.Person, error) {
	query := "select * from people where id=?"
	stmt, _ := db.PrepareStatement(query)
	r := stmt.QueryRow(id)
	return parse(r)
}

func (repo *PeopleRepo) FindByEmail(email string) (*models.Person, error) {
	query := "select * from people where email=?"
	stmt, _ := db.PrepareStatement(query)
	r := stmt.QueryRow(email)
	return parse(r)
}

func (repo *PeopleRepo) Update(p *models.Person) (int64, error) {
	query := "update people set name=?, surname=?, email=?, phone=?, updatedAt=? where id=?"
	stmt, _ := db.PrepareStatement(query)
	timestamp := time.Now()
	r, err := stmt.Exec(p.Name, p.Surname, p.Email, p.Phone, timestamp, p.Id)
	if err != nil {
		return 0, err
	}
	p.UpdatedAt = &timestamp
	return r.RowsAffected()
}

func (repo *PeopleRepo) Delete(p *models.Person) (int64, error) {
	query := "update people set deletedAt=? where id=?"
	stmt, _ := db.PrepareStatement(query)
	timestamp := time.Now()
	r, err := stmt.Exec(timestamp, p.Id)
	if err != nil {
		return 0, err
	}
	p.DeletedAt = &timestamp
	return r.RowsAffected()
}

func parse(r *sql.Row) (*models.Person, error) {
	p := &models.Person{}
	r.Scan(&p.Id,
		&p.Name,
		&p.Surname,
		&p.Email,
		&p.Phone,
		&p.CreatedAt, &p.UpdatedAt, &p.DeletedAt)
	if p.CreatedAt == nil {
		p = nil
	}
	return p, nil
}
