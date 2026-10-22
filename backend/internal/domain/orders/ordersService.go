package orders

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"example/internal/api"
	"example/internal/domain/addresses"
	"example/internal/domain/products"

	"github.com/jackc/pgx/v5"
	"uuid"
)

type service struct {
	*repo
	productsService  *products.Service
	addressesService *addresses.Service
}

func NewService(repo *repo, productsService *products.Service, addressesService *addresses.Service) *service {
	return &service{repo, productsService, addressesService}
}

func (this *service) Create(ctx context.Context, order *Model) *api.Error {
	order.Id = uuid.NewV7().String()
	// Toda orden entra por el mismo lado: recién creada y pendiente.
	order.Status = StatusPending
	order.Total = 0 // se calcula al leer, jamás del body

	// La dirección tiene que existir y ser del que pide (FindById responde
	// 404 si no es suya): si no, cualquiera podría leer la dirección de otro
	// en su propio historial. La que queda en la orden es la de la DB.
	if order.Address == nil || order.Address.Id == "" {
		return api.NewError("La orden no tiene dirección.", http.StatusBadRequest)
	}
	address, err := this.addressesService.FindById(ctx, order.Address.Id)
	if err != nil {
		return err
	}
	if address == nil {
		return api.NewError("La dirección no es válida.", http.StatusBadRequest)
	}
	order.Address = address

	// Los platos tienen que existir, con cantidad válida y sin repetir: si
	// no, el insert reviente contra `ppu not null` o contra la PK (500) en
	// vez de responder 400.
	if len(order.Products) == 0 {
		return api.NewError("La orden no tiene platos.", http.StatusBadRequest)
	}
	seen := map[string]bool{}
	for _, product := range order.Products {
		if product == nil || product.Id == "" {
			return api.NewError("Plato inválido.", http.StatusBadRequest)
		}
		if product.Amount < 1 {
			return api.NewError("La cantidad de un plato no puede ser menor a 1.", http.StatusBadRequest)
		}
		if seen[product.Id] {
			return api.NewError("No se puede repetir el mismo plato.", http.StatusBadRequest)
		}
		seen[product.Id] = true
		// 404 si no existe (o si está borrado): no se pueden pedir fantasmas.
		if _, err := this.productsService.FindById(ctx, product.Id); err != nil {
			return err
		}
	}

	if err := this.repo.create(ctx, order); err != nil {
		return api.NewGenericError(err)
	}
	return nil
}

func (this *service) FindById(ctx context.Context, id string) (*Model, *api.Error) {
	order, err := this.repo.findById(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil // no existe ≠ error de infraestructura
	}
	if err != nil {
		return nil, api.NewGenericError(err)
	}
	return order, nil
}

func (this *service) FindByUser(ctx context.Context, userID string, page, size int) ([]*Model, int, *api.Error) {
	orders, pages, err := this.repo.findByuserID(ctx, userID, page, size)
	if err != nil {
		return nil, 0, api.NewGenericError(err)
	}
	return orders, pages, nil
}

// FindAll es la vista del admin: todas las órdenes, con filtros opcionales de
// estado y de fechas ("AAAA-MM-DD"; desde y hasta inclusive).
func (this *service) FindAll(ctx context.Context, status, from, to string, page, size int) ([]*Model, int, *api.Error) {
	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 10
	}

	status = strings.ToLower(strings.TrimSpace(status))
	if status != "" && !ValidStatus(status) {
		return nil, 0, api.NewError("Estado no válido.", http.StatusBadRequest)
	}

	var fromArg, toArg any // nil → sin filtro
	if from != "" {
		day, err := parseDay("desde", from)
		if err != nil {
			return nil, 0, err
		}
		fromArg = day
	}
	if to != "" {
		day, err := parseDay("hasta", to)
		if err != nil {
			return nil, 0, err
		}
		toArg = day.AddDate(0, 0, 1) // el hasta es inclusive: se busca < al día siguiente
	}

	orders, pages, err := this.repo.listAll(ctx, status, fromArg, toArg, page, size)
	if err != nil {
		return nil, 0, api.NewGenericError(err)
	}
	return orders, pages, nil
}

// parseDay convierte "AAAA-MM-DD" a timestamp para filtrar.
func parseDay(label, value string) (time.Time, *api.Error) {
	day, err := time.Parse(time.DateOnly, value)
	if err != nil {
		return time.Time{}, api.NewError(
			"Fecha inválida ("+label+"): use AAAA-MM-DD.", http.StatusBadRequest)
	}
	return day, nil
}

func (this *service) UpdateStatus(ctx context.Context, id, status string) *api.Error {
	if !ValidStatus(status) {
		return api.NewError("Estado no válido.", http.StatusBadRequest)
	}

	updated, err := this.repo.updateStatus(ctx, id, status)
	if err != nil {
		return api.NewGenericError(err)
	}
	if !updated {
		return api.NewError("Orden no encontrada.", http.StatusNotFound)
	}
	return nil
}
