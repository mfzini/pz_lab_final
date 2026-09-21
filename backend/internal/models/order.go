package models

type Order struct {
	Id,
	Detail,
	Status string
	Client *Person
	Dishes []*Dish
	TimeStamps
}
