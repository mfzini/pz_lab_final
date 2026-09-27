package dishes

import (
	"context"
	"example/internal/apiErrors"
	"net/http"
	"time"
	"uuid"
)

type service struct {
	repo *repo
}

func NewService(repo *repo) *service {
	return &service{repo}
}

func (this *service) Create(ctx context.Context, dish *Model) *apiErrors.Error {
	if dish.Price < 0 {
		return apiErrors.New("Estas perdiendo plata!", http.StatusBadRequest)
	}

	timestamp := time.Now()
	dish.Id = uuid.NewV7().String()
	dish.CreatedAt = &timestamp
	dish.Portrait.CreatedAt = &timestamp
	if err := this.repo.create(ctx, dish); err != nil {
		return apiErrors.Generic(err)
	}
	return nil
}
