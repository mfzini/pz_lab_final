package orders

import (
	"context"
	"example/internal/apiErrors"
	"example/internal/domain/products"
)

type service struct {
	*repo
	productsService *products.Service
}

func NewService(repo *repo, productsService *products.Service) *service {
	return &service{repo, productsService}
}

func (this *service) create(ctx context.Context, order *Model) *apiErrors.Error {
	if err := this.repo.create(ctx, order); err != nil {
		return apiErrors.Generic(err)
	}

	return nil
}
