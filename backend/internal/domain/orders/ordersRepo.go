package orders

import (
	"context"
	"example/internal/domain/addresses"
	"example/internal/domain/people"
	"example/internal/domain/products"

	"github.com/jackc/pgx/v5"
)

type repo struct {
	db *pgx.Conn
}

func NewRepo(connection *pgx.Conn) *repo {
	return &repo{db: connection}
}

func (this *repo) create(ctx context.Context, order *Model) error {
	tx, err := this.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	batch := &pgx.Batch{}
	query := "insert into orders (id, client_id, address_id, detail) values ($1, $2, $3, $4);"
	batch.Queue(query, order.Id, order.Client.Id, order.Address.Id, order.Detail)

	for _, product := range order.Products {
		query := `insert into orders_products (order_id, product_id, amount, ppu) values
			($1, $2, $3, (select price from products where id = $2));`
		batch.Queue(query, order.Id, product.Id, product.Amount)
	}

	if err := tx.SendBatch(ctx, batch).Close(); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (this *repo) findById(ctx context.Context, id string) (*Model, error) {
	query := `
		select o.id, o.client_id, o.address_id, o.detail
			person.name, person.surname, person.phone, person.email
			a.street, a.house, a.door, a.floor, a.detail, a.coordinates::text as coordinates
		from orders o
		join people p on person.id = o.client_id
		join addresses a on a.id = o.address_id
		where o.id = $1
	`
	row := this.db.QueryRow(ctx, query, id)
	order := &Model{Total: 0}
	person := &people.Model{}
	address := &addresses.Model{}
	err := row.Scan(&order.Id, &person.Id, &address.Id, &order.Detail,
		&person.Name, &person.Surname, &person.Phone, &person.Email,
		&address.Street, &address.House, &address.Door, &address.Floor, &address.Detail, &address.Coordinates)
	if err != nil {
		return nil, err
	}
	order.Client = person
	order.Address = address
	if err := this.fetchProducts(ctx, order); err != nil {
		return nil, err
	}
	return order, nil
}
func (this *repo) findByPersonID(ctx context.Context, id string) ([]*Model, error) {
	query := `
		select o.id, o.client_id, o.address_id, o.detail
			person.name, person.surname, person.phone, person.email
			a.street, a.house, a.door, a.floor, a.detail, a.coordinates::text as coordinates
		from orders o
		join people p on person.id = o.client_id
		join addresses a on a.id = o.address_id
		where o.client_id = $1
	`
	rows, err := this.db.Query(ctx, query, id)
	if err != nil {
		return nil, err
	}
	orders := make([]*Model, 0)
	person := &people.Model{}
	for rows.Next() {
		order := &Model{Total: 0}
		address := &addresses.Model{}
		err := rows.Scan(&order.Id, &person.Id, &address.Id, &order.Detail,
			&person.Name, &person.Surname, &person.Phone, &person.Email,
			&address.Street, &address.House, &address.Door, &address.Floor, &address.Detail, &address.Coordinates)
		if err != nil {
			return nil, err
		}
		order.Client = person
		order.Address = address
		if err := this.fetchProducts(ctx, order); err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}
	return orders, nil
}

func (this *repo) fetchProducts(ctx context.Context, order *Model) error {
	query := `
		select d.id, d.name, od.amount, od.ppu
		from orders_products od
		join products d on d.id = od.product_id
		where od.order_id = $1
	`
	rows, err := this.db.Query(ctx, query, order.Id)
	if err != nil {
		return err
	}

	order.Products = make([]*products.Model, 0)
	for rows.Next() {
		product := &products.Model{}
		if err := rows.Scan(&product.Id, &product.Name, &product.Amount, &product.Price); err != nil {
			return err
		}
		product.Price = product.Price * float64(product.Amount)
		order.Products = append(order.Products, product)
		order.Total += product.Price
	}
	return nil
}
