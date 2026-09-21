package models

type Dish struct {
	Id,
	Name string
	Ingredients []Ingredient
	Price       float64
	TimeStamps
}
