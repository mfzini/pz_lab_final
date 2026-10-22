package categories

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"example/internal/api"

	"github.com/jackc/pgx/v5/pgconn"
)

type Service struct {
	repo *repo
}

func NewService(repo *repo) *Service {
	return &Service{repo}
}

func (this *Service) Create(ctx context.Context, category *Model) *api.Error {
	category.Name = strings.TrimSpace(category.Name)
	if category.Name == "" {
		return api.NewError("Falta el nombre de la categoría.", http.StatusBadRequest)
	}
	if err := this.repo.create(ctx, category); err != nil {
		// 23505 = unique_violation sobre categories.name.
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return api.NewError("Ya existe una categoría con ese nombre.", http.StatusConflict)
		}
		return api.NewGenericError(err)
	}
	return nil
}

// List con filtro opcional de nombre (búsqueda asíncrona del form de platos).
func (this *Service) List(ctx context.Context, name string, page, size int) ([]*Model, int, *api.Error) {
	categories, pages, err := this.repo.list(ctx, name, page, size)
	if err != nil {
		return nil, 0, api.NewGenericError(err)
	}
	return categories, pages, nil
}
