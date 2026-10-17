package products

import (
	"context"
	"example/internal/api"
	"example/internal/domain/images"
	"fmt"
	"mime/multipart"
	"net/http"
	"uuid"
)

type Service struct {
	repo          *repo
	imagesService *images.Service
}

func NewService(repo *repo, imagesService *images.Service) *Service {
	return &Service{repo, imagesService}
}

func (this *Service) Create(ctx context.Context, product *Model, img multipart.File) *api.Error {
	if product.Price <= 0 {
		return api.NewError("Estas perdiendo plata!", http.StatusBadRequest)
	}
	hash, err := this.imagesService.Save(img)
	if err != nil {
		return api.NewGenericError(err)
	}
	product.Id = uuid.NewV7().String()
	product.Portrait = fmt.Sprintf("/uploads/%s", hash)
	if err := this.repo.create(ctx, product); err != nil {
		return api.NewGenericError(err)
	}
	return nil
}

func (this *Service) Update(ctx context.Context, product *Model, img multipart.File) *api.Error {
	if product.Price <= 0 {
		return api.NewError("Estas perdiendo plata!", http.StatusBadRequest)
	}
	hash, err := this.imagesService.Save(img)
	if err != nil {
		return api.NewGenericError(err)
	}
	product.Portrait = fmt.Sprintf("/uploads/%s", hash)
	if err := this.repo.update(ctx, product); err != nil {
		return api.NewGenericError(err)
	}
	return nil
}

func (this *Service) FindById(ctx context.Context, id string) (*Model, error) {
	dish, err := this.repo.findById(ctx, id)
	if err != nil {
		return nil, err
	}
	return dish, nil
}

func (this *Service) List(ctx context.Context, name, sortBy string, page, size int) ([]*Model, int, error) {
	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 10
	}
	sorter := PRODUCT_SORTER(sortBy)
	switch sorter {
	case SORT_BY_CREATION, SORT_BY_NAME, SORT_BY_PRICE:
		break
	default:
		sorter = SORT_BY_NAME
	}

	if name != "" {
		return this.repo.listByName(ctx, name, page, size, sorter)
	} else {
		return this.repo.list(ctx, page, size, sorter)
	}
}
