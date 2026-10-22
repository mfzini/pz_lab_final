// Command seeder carga datos de prueba sobre un esquema ya creado con
// cmd/seeder/script_pg.sql:
//
//	cd backend && go run ./cmd/seeder
//
// Se puede re-correr sin problemas: usa ids fijos y "on conflict do nothing",
// así que no duplica nada (lo que ya estaba, se saltea).
package main

import (
	"context"
	"fmt"
	"os"

	"example/internal/db"

	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

// Contraseña de los tres usuarios de prueba (se hashea al correr el seeder).
const password = "clave1234"

// Ids fijos: son los que usan los inserts y los "on conflict".
const (
	adminID        = "00000000-0000-4000-8000-000000000001"
	juanID         = "00000000-0000-4000-8000-000000000002"
	mariaID        = "00000000-0000-4000-8000-000000000003"
	juanAddr       = "00000000-0000-4000-8000-000000000101"
	mariaAddr      = "00000000-0000-4000-8000-000000000102"
	hamburguesa    = "00000000-0000-4000-8000-000000000201"
	papas          = "00000000-0000-4000-8000-000000000202"
	pizza          = "00000000-0000-4000-8000-000000000203"
	ensalada       = "00000000-0000-4000-8000-000000000204"
	gaseosa        = "00000000-0000-4000-8000-000000000205"
	flan           = "00000000-0000-4000-8000-000000000206"
	orderShipped   = "00000000-0000-4000-8000-000000000301"
	orderPreparing = "00000000-0000-4000-8000-000000000302"
	orderPending   = "00000000-0000-4000-8000-000000000303"
	catComida      = "00000000-0000-4000-8000-000000000211"
	catBebidas     = "00000000-0000-4000-8000-000000000212"
	catPostres     = "00000000-0000-4000-8000-000000000213"
)

func main() {
	ctx := context.Background()
	conn := db.GetConnection()
	defer conn.Close()

	hash, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	if err != nil {
		die(err)
	}

	total := 0

	// --- Usuarios: un admin y dos clientes (más adelante sirve para probar
	// que un usuario no ve los datos del otro). María se queda sin teléfono:
	// phone es anulable y el modelo lo lee como puntero.
	total += exec(ctx, conn, `insert into users (id, name, surname, email, phone, password, role)
		values ($1, 'Cosme', 'Fulanito', 'admin@test.com', '1100000001', $2, 'admin')
		on conflict do nothing`, adminID, hash)
	total += exec(ctx, conn, `insert into users (id, name, surname, email, phone, password, role)
		values ($1, 'Juan', 'Perez', 'juan@test.com', '1100000002', $2, 'client')
		on conflict do nothing`, juanID, hash)
	total += exec(ctx, conn, `insert into users (id, name, surname, email, password, role)
		values ($1, 'Maria', 'Gomez', 'maria@test.com', $2, 'client')
		on conflict do nothing`, mariaID, hash)

	// --- Una dirección por cliente ("12B" y "742" prueban house como texto).
	// La de María queda sin coordenadas (point anulable).
	total += exec(ctx, conn, `insert into addresses (id, user_id, street, house, floor, door, coordinates, detail)
		values ($1, $2, 'Av. Siempreviva', '742', 1, 'A', point(-58.38, -34.60), 'Timbre roto')
		on conflict do nothing`, juanAddr, juanID)
	total += exec(ctx, conn, `insert into addresses (id, user_id, street, house)
		values ($1, $2, 'Calle Falsa', '12B')
		on conflict do nothing`, mariaAddr, mariaID)

	// --- Tres categorías (maestra chica): los platos de abajo se asignan
	// a una de ellas (products N:1 categories).
	for _, c := range []struct{ id, name string }{
		{catComida, "Comida"},
		{catBebidas, "Bebidas"},
		{catPostres, "Postres"},
	} {
		total += exec(ctx, conn, `insert into categories (id, name)
			values ($1, $2) on conflict do nothing`, c.id, c.name)
	}

	// --- Seis platos, precios en centavos (1250 = $12,50). Sin foto:
	// portrait queda en null y las lecturas no tienen que romperse.
	for _, p := range []struct {
		id, name string
		price    int
	}{
		{hamburguesa, "Hamburguesa Clasica", 1250},
		{papas, "Papas con Cheddar", 900},
		{pizza, "Pizza Muzzarella", 1500},
		{ensalada, "Ensalada Mixta", 850},
		{gaseosa, "Gaseosa 500cc", 500},
		{flan, "Flan Casero", 700},
	} {
		total += exec(ctx, conn, `insert into products (id, name, price)
			values ($1, $2, $3) on conflict do nothing`, p.id, p.name, p.price)
	}

	// --- Categoría de cada plato. El "where category_id is null" hace que
	// el seeder se pueda re-correr sin pisar lo que el admin cambió.
	for _, a := range []struct{ product, category string }{
		{hamburguesa, catComida},
		{papas, catComida},
		{pizza, catComida},
		{ensalada, catComida},
		{gaseosa, catBebidas},
		{flan, catPostres},
	} {
		total += exec(ctx, conn, `update products set category_id = $2
			where id = $1 and category_id is null`, a.product, a.category)
	}

	// --- Tres pedidos, uno por estado y repartidos en el tiempo: así se
	// pueden probar los filtros de estado y de fecha de GET /ordenes/admin.
	// El detalle cubre los tres casos: texto ('Sin cebolla'), null (acá) y ''.
	total += exec(ctx, conn, `insert into orders (id, client_id, address_id, detail, status, created_at, updated_at)
		values ($1, $2, $3, 'Sin cebolla', 'shipped',
			now() - interval '5 days', now() - interval '4 days')
		on conflict do nothing`, orderShipped, juanID, juanAddr)
	total += exec(ctx, conn, `insert into orders (id, client_id, address_id, status, created_at, updated_at)
		values ($1, $2, $3, 'preparing',
			now() - interval '2 days', now() - interval '1 day')
		on conflict do nothing`, orderPreparing, juanID, juanAddr)
	total += exec(ctx, conn, `insert into orders (id, client_id, address_id, detail, status)
		values ($1, $2, $3, '', 'pending')
		on conflict do nothing`, orderPending, mariaID, mariaAddr)

	// --- Líneas de cada pedido: el ppu se fija igual que hace el repo, con
	// el precio del plato en ese momento (acá, el de la semilla).
	for _, l := range []struct {
		orderID, productID string
		amount, ppu        int
	}{
		{orderShipped, hamburguesa, 2, 1250},
		{orderShipped, gaseosa, 2, 500},
		{orderPreparing, pizza, 1, 1500},
		{orderPreparing, papas, 1, 900},
		{orderPending, ensalada, 1, 850},
		{orderPending, flan, 2, 700},
	} {
		total += exec(ctx, conn, `insert into orders_products (order_id, product_id, amount, ppu)
			values ($1, $2, $3, $4) on conflict do nothing`, l.orderID, l.productID, l.amount, l.ppu)
	}

	fmt.Printf("Datos de prueba cargados: %d filas nuevas.\n", total)
	fmt.Printf("Usuarios (contraseña %q): admin@test.com, juan@test.com, maria@test.com\n", password)
}

func exec(ctx context.Context, conn *pgxpool.Pool, query string, args ...any) int {
	tag, err := conn.Exec(ctx, query, args...)
	if err != nil {
		fmt.Fprintf(os.Stderr, "no se pudo insertar: %v\n", err)
		fmt.Fprintln(os.Stderr, "¿Corriste primero backend/cmd/seeder/script_pg.sql?")
		os.Exit(1)
	}
	return int(tag.RowsAffected())
}

func die(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
