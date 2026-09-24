package services

import (
	"context"
	"example/internal/models"
	"example/internal/repositories"
)

type DishesService struct {
	Repo *repositories.DishesRepo
}

func (self *DishesService) Create(ctx context.Context, dish *models.Dish) {
	// aca deberiamos hacer validaciones
	self.Repo.Create(ctx, dish)
}
