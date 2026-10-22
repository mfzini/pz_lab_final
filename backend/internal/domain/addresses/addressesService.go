package addresses

import (
	"context"
	"errors"
	"net/http"

	"example/internal/api"
	"example/internal/security"

	"github.com/jackc/pgx/v5"
	"uuid"
)

// Service es el servicio de direcciones: además del CRUD aplica ownership
// (dueño o admin) en las lecturas y modificaciones.
type Service struct {
	repo *repo
}

func NewService(r *repo) *Service {
	return &Service{repo: r}
}

func (this *Service) Create(ctx context.Context, address *Model) *api.Error {
	claims := security.GetClaimsFromContext(ctx)
	if claims == nil {
		return api.NewError("Se requiere autenticación.", http.StatusUnauthorized)
	}

	address.Id = uuid.NewV7().String()
	address.User_ID = claims.Subject // el dueño sale del token, nunca del body

	if err := this.repo.create(ctx, address); err != nil {
		return api.NewGenericError(err)
	}
	return nil
}

func (this *Service) FindById(ctx context.Context, id string) (*Model, *api.Error) {
	address, err := this.repo.FindById(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, api.NewGenericError(err)
	}

	claims := security.GetClaimsFromContext(ctx)
	if claims == nil {
		return nil, api.NewError("Se requiere autenticación.", http.StatusUnauthorized)
	}
	// Dueño o admin. Se responde 404 en vez de 403 para no filtrar que existe.
	if claims.Subject != address.User_ID && claims.Role != security.Admin {
		return nil, api.NewError("Dirección no encontrada.", http.StatusNotFound)
	}
	return address, nil
}

// FindByUserID con filtro opcional de calle (street = "" lista todo).
func (this *Service) FindByUserID(ctx context.Context, id, street string, page, size int) ([]*Model, int, *api.Error) {
	addresses, pages, err := this.repo.FindByUserID(ctx, id, street, page, size)
	if err != nil {
		return nil, 0, api.NewGenericError(err)
	}
	return addresses, pages, nil
}

func (this *Service) FindByStreetAndHouse(ctx context.Context, street, house string, page, size int) ([]*Model, int, *api.Error) {
	addresses, pages, err := this.repo.FindByStreetAndHouse(ctx, street, house, page, size)
	if err != nil {
		return nil, 0, api.NewGenericError(err)
	}
	return addresses, pages, nil
}

func (this *Service) Update(ctx context.Context, id string, data *Model) *api.Error {
	address, err := this.repo.FindById(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return api.NewError("Dirección no encontrada.", http.StatusNotFound)
	}
	if err != nil {
		return api.NewGenericError(err)
	}

	claims := security.GetClaimsFromContext(ctx)
	if claims == nil {
		return api.NewError("Se requiere autenticación.", http.StatusUnauthorized)
	}
	// La comparación es con el dueño real (el de la DB), no con lo que diga el body.
	if claims.Subject != address.User_ID && claims.Role != security.Admin {
		return api.NewError("No podés modificar esta dirección.", http.StatusForbidden)
	}

	data.User_ID = address.User_ID // el dueño no cambia
	data.Id = address.Id           // el id sale del path, no del body

	if err := this.repo.update(ctx, data); err != nil {
		return api.NewGenericError(err)
	}
	return nil
}

func (this *Service) Delete(ctx context.Context, id string) *api.Error {
	// FindById ya valida dueño o admin.
	address, err := this.FindById(ctx, id)
	if err != nil {
		return err
	}
	if address == nil {
		return api.NewError("Dirección no encontrada.", http.StatusNotFound)
	}

	if err := this.repo.delete(ctx, address); err != nil {
		return api.NewGenericError(err)
	}
	return nil
}
