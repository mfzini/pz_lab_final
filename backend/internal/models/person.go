package models

import "uuid"

type Person struct {
	Id uuid.UUID
	Name,
	Surname,
	Phone,
	Email string
	TimeStamps
}

func NewPerson(name, surname, email, phone string) *Person {
	return &Person{
		Id:      uuid.NewV7(),
		Name:    name,
		Surname: surname,
		Email:   email,
		Phone:   phone,
	}
}
