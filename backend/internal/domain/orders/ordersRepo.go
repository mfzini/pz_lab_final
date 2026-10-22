package orders

import (
	"context"
	"errors"

	"example/internal/domain/addresses"
	"example/internal/domain/common"
	"example/internal/domain/products"
	"example/internal/domain/users"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type repo struct {
	db *pgxpool.Pool
}

func NewRepo(connection *pgxpool.Pool) *repo {
	return &repo{db: connection}
}

func (this *repo) create(ctx context.Context, order *Model) error {
	if order.Client == nil || order.Client.Id == "" {
		return errors.New("la orden no tiene cliente")
	}
	if order.Address == nil || order.Address.Id == "" {
		return errors.New("la orden no tiene dirección")
	}

	tx, err := this.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	batch := &pgx.Batch{}
	batch.Queue(
		`insert into orders (id, client_id, address_id, detail, status)
		values ($1, $2, $3, $4, $5);`,
		order.Id, order.Client.Id, order.Address.Id, order.Detail, order.Status,
	)

	for _, product := range order.Products {
		batch.Queue(
			`insert into orders_products (order_id, product_id, amount, ppu) values
			($1, $2, $3, (select price from products where id = $2));`,
			order.Id, product.Id, product.Amount,
		)
	}

	if err := tx.SendBatch(ctx, batch).Close(); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (this *repo) findById(ctx context.Context, id string) (*Model, error) {
	query := `
		select o.id, o.client_id, o.address_id, o.detail, o.status,
			u.name, u.surname, u.phone, u.email,
			a.street, a.house, a.door, a.floor, a.detail, a.coordinates::text as coordinates
		from orders o
		join users u on u.id = o.client_id
		join addresses a on a.id = o.address_id
		where o.id = $1
	`
	row := this.db.QueryRow(ctx, query, id)
	order := &Model{Total: 0}
	user := &users.Model{}
	address := &addresses.Model{}
	err := row.Scan(&order.Id, &user.Id, &address.Id, &order.Detail, &order.Status,
		&user.Name, &user.Surname, &user.Phone, &user.Email,
		&address.Street, &address.House, &address.Door, &address.Floor, &address.Detail, &address.Coordinates)
	if err != nil {
		return nil, err
	}
	order.Client = user
	order.Address = address
	if err := this.fetchProducts(ctx, order); err != nil {
		return nil, err
	}
	return order, nil
}

// findByuserID es el historial del usuario: página de órdenes + total de páginas.
func (this *repo) findByuserID(ctx context.Context, id string, page, size int) ([]*Model, int, error) {
	offset := (page - 1) * size
	query := `
		select o.id, o.client_id, o.address_id, o.detail, o.status,
			u.name, u.surname, u.phone, u.email,
			a.street, a.house, a.door, a.floor, a.detail, a.coordinates::text as coordinates
		from orders o
		join users u on u.id = o.client_id
		join addresses a on a.id = o.address_id
		where o.client_id = $1
		order by o.created_at desc, o.id
		limit $2 offset $3
	`
	rows, err := this.db.Query(ctx, query, id, size, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	orders := make([]*Model, 0)
	// Todas las órdenes son del mismo cliente: un solo user compartido alcanza.
	user := &users.Model{}
	for rows.Next() {
		order := &Model{Total: 0}
		address := &addresses.Model{}
		err := rows.Scan(&order.Id, &user.Id, &address.Id, &order.Detail, &order.Status,
			&user.Name, &user.Surname, &user.Phone, &user.Email,
			&address.Street, &address.House, &address.Door, &address.Floor, &address.Detail, &address.Coordinates)
		if err != nil {
			return nil, 0, err
		}
		order.Client = user
		order.Address = address
		orders = append(orders, order)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	// Los productos se buscan recién acá, con el result set ya cerrado: la
	// conexión es única y no admite dos queries al mismo tiempo.
	for _, order := range orders {
		if err := this.fetchProducts(ctx, order); err != nil {
			return nil, 0, err
		}
	}

	totalRows := 0
	if err := this.db.QueryRow(ctx, "select count(*) from orders where client_id = $1", id).Scan(&totalRows); err != nil {
		return nil, 0, err
	}
	return orders, common.TotalPagesFor(totalRows, size), nil
}

// listAll devuelve todas las órdenes (cualquier cliente), filtradas por estado
// y rango de fechas si vienen ($2/$3 = nil → sin filtro), más recientes primero.
func (this *repo) listAll(ctx context.Context, status string, from, to any, page, size int) ([]*Model, int, error) {
	// Los casts de $2/$3 no son decorativos: con "$2 is null" solo, Postgres
	// no puede inferir el tipo del parámetro (SQLSTATE 42P08) cuando llega
	// sin tipar (que es justo lo que pasa cuando no hay filtro de fecha).
	where := `
		where ($1 = '' or o.status = $1)
			and ($2::timestamptz is null or o.created_at >= $2)
			and ($3::timestamptz is null or o.created_at < $3)
	`
	query := `
		select o.id, o.client_id, o.address_id, o.detail, o.status,
			u.name, u.surname, u.phone, u.email,
			a.street, a.house, a.door, a.floor, a.detail, a.coordinates::text as coordinates
		from orders o
		join users u on u.id = o.client_id
		join addresses a on a.id = o.address_id
	` + where + `
		order by o.created_at desc, o.id
		limit $4 offset $5
	`
	offset := (page - 1) * size
	rows, err := this.db.Query(ctx, query, status, from, to, size, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	orders := make([]*Model, 0)
	for rows.Next() {
		// Acá cada orden puede ser de un cliente distinto: user propio por fila.
		order := &Model{Total: 0}
		user := &users.Model{}
		address := &addresses.Model{}
		if err := rows.Scan(&order.Id, &user.Id, &address.Id, &order.Detail, &order.Status,
			&user.Name, &user.Surname, &user.Phone, &user.Email,
			&address.Street, &address.House, &address.Door, &address.Floor, &address.Detail, &address.Coordinates); err != nil {
			return nil, 0, err
		}
		order.Client = user
		order.Address = address
		orders = append(orders, order)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	// Igual que en findByuserID: los productos, con el result set cerrado.
	for _, order := range orders {
		if err := this.fetchProducts(ctx, order); err != nil {
			return nil, 0, err
		}
	}

	countQuery := "select count(*) from orders o " + where
	totalRows := 0
	if err := this.db.QueryRow(ctx, countQuery, status, from, to).Scan(&totalRows); err != nil {
		return nil, 0, err
	}
	return orders, common.TotalPagesFor(totalRows, size), nil
}

// updateStatus cambia el estado y devuelve false si la orden no existía.
func (this *repo) updateStatus(ctx context.Context, id, status string) (bool, error) {
	query := "update orders set status = $1, updated_at = now() where id = $2"
	tag, err := this.db.Exec(ctx, query, status, id)
	if err != nil {
		return false, err
	}
	return tag.RowsAffected() > 0, nil
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
		// ppu viene en centavos: el importe de la línea y el total quedan en centavos.
		product.Price = product.Price * product.Amount
		order.Products = append(order.Products, product)
		order.Total += product.Price
	}
	return nil
}
