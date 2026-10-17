package addresses

import (
	"context"
	"example/internal/api"
	"example/internal/security"
	"net/http"
)

type service struct {
	*repo
}

func NewService(repo *repo) *service {
	return &service{repo}
}

func (this *service) Create(ctx context.Context, address *Model) *api.Error {
	this.repo.create(ctx, address)
	return nil
}

func (this *service) Update(ctx context.Context, id string, data *Model) *api.Error {
	claims := security.GetClaimsFromContext(ctx)
	address, err := this.repo.FindById(ctx, id)
	if err != nil {
		return api.NewGenericError(err)
	}
	if address == nil {
		api.NewError("Address not found", http.StatusNotFound)
	}
	if claims.Subject != data.User_ID || claims.Role != security.Admin {
		return api.NewError("Not allowed.", http.StatusForbidden)
	}

	data.User_ID = address.User_ID
	data.ID = address.ID

	if err := this.repo.update(ctx, data); err != nil {
		return api.NewGenericError(err)
	}
	return nil
}

func (this *service) Delete(ctx context.Context, id string) *api.Error {
	claims := security.GetClaimsFromContext(ctx)
	address, err := this.FindById(ctx, id)
	if err != nil {
		return err
	}
	if claims.Subject != address.User_ID || claims.Role != security.Admin {
		return api.NewError("Not allowed.", http.StatusForbidden)
	}
	if err := this.repo.delete(ctx, address); err != nil {
		return api.NewGenericError(err)
	}
	return nil
}

func (this *service) FindById(ctx context.Context, id string) (*Model, *api.Error) {
	addresses, err := this.repo.FindById(ctx, id)
	if err != nil {
		return nil, api.NewError(err.Error(), http.StatusInternalServerError)
	}
	if addresses == nil {
		return nil, api.NewError("Address not found.", http.StatusNotFound)
	}
	claims := security.GetClaimsFromContext(ctx)
	if claims.Subject != addresses.User_ID || claims.Role != security.Admin {
		return nil, api.NewError("Address not found.", http.StatusNotFound)
	}
	return addresses, nil
}

func (this *service) FindByUserID(ctx context.Context, id string) ([]*Model, *api.Error) {
	addresses, err := this.repo.FindByUserID(ctx, id)
	if err != nil {
		return nil, api.NewGenericError(err)
	}
	return addresses, nil
}
