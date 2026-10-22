package orders

import (
	"context"
	"os"
	"testing"
	"time"

	"example/internal/db"
)

// Los ids de la semilla (cmd/seeder): el pedido "preparing" no tiene detalle,
// y el de María une un usuario sin teléfono con una dirección sin coordenadas.
const (
	testOrderPreparing = "00000000-0000-4000-8000-000000000302"
	testOrderShipped   = "00000000-0000-4000-8000-000000000301"
	testOrderPending   = "00000000-0000-4000-8000-000000000303"
	testMariaID        = "00000000-0000-4000-8000-000000000003"
)

// Verifica que leer columnas anulables en null no reviente: los modelos usan
// puntero para esas columnas, y los scans manuales de este repo pasan por
// puntero doble. Corre contra la DB real; sin env, se saltea.
func TestReadsTolerateNulls(t *testing.T) {
	if os.Getenv("POSTGRES_HOST") == "" {
		t.Skip("sin POSTGRES_* en el entorno")
	}
	ctx := context.Background()
	repo := NewRepo(db.GetConnection())

	// findById: detail null.
	order, err := repo.findById(ctx, testOrderPreparing)
	if err != nil {
		t.Fatalf("findById con detail null falló: %v", err)
	}
	if order.Detail != nil {
		t.Errorf("esperaba detail null, obtuve %q", *order.Detail)
	}

	// findByuserID: phone null (usuario) y coordinates null (dirección).
	// Cualquier pedido de María usa su única dirección (sin coordenadas), así
	// que basta con que haya al menos uno.
	orders, _, err := repo.findByuserID(ctx, testMariaID, 1, 100)
	if err != nil {
		t.Fatalf("findByuserID con phone/coordinates null falló: %v", err)
	}
	if len(orders) < 1 {
		t.Fatalf("esperaba al menos 1 pedido de María, obtuve %d", len(orders))
	}
	if orders[0].Client.Phone != nil {
		t.Errorf("esperaba phone null, obtuve %q", *orders[0].Client.Phone)
	}
	if orders[0].Address.Coordinates != nil {
		t.Errorf("esperaba coordinates null, obtuve %q", *orders[0].Address.Coordinates)
	}

	// listAll: el mismo barrido con todos los pedidos. La DB puede tener más
	// que la semilla (pruebas manuales del checkout), así que sólo exigimos
	// los 3 originales y al menos una página.
	all, pages, err := repo.listAll(ctx, "", nil, nil, 1, 100)
	if err != nil {
		t.Fatalf("listAll con nulls falló: %v", err)
	}
	if len(all) < 3 || pages < 1 {
		t.Fatalf("esperaba al menos los 3 pedidos de la semilla en %d página(s), obtuve %d", pages, len(all))
	}
}

// Los filtros del admin: el camino en que $2/$3 SÍ llevan fecha (además del
// nil del test anterior). Usa pedidos propios con estado y fecha conocidos:
// la data de la semilla la puede cambiar el admin (estados) o el cliente
// (pedidos nuevos), y el test tiene que seguir decidiendo.
func TestListAllFilters(t *testing.T) {
	if os.Getenv("POSTGRES_HOST") == "" {
		t.Skip("sin POSTGRES_* en el entorno")
	}
	ctx := context.Background()
	repo := NewRepo(db.GetConnection())
	conn := db.GetConnection()
	now := time.Now()

	const (
		tmpPending = "cccccccc-0000-4000-8000-000000000001"
		tmpShipped = "cccccccc-0000-4000-8000-000000000002"
		clientID   = "00000000-0000-4000-8000-000000000003" // María (semilla)
		addrID     = "00000000-0000-4000-8000-000000000102" // dirección de María
	)

	if _, err := conn.Exec(ctx, `delete from orders where id in ($1, $2)`, tmpPending, tmpShipped); err != nil {
		t.Fatalf("limpieza previa: %v", err)
	}
	if _, err := conn.Exec(ctx, `insert into orders (id, client_id, address_id, status, created_at)
		values ($1, $3, $4, 'pending', now() - interval '1 day'),
		       ($2, $3, $4, 'shipped', now() - interval '1 hour')`,
		tmpPending, tmpShipped, clientID, addrID); err != nil {
		t.Fatalf("insert de prueba: %v", err)
	}
	defer conn.Exec(ctx, `delete from orders where id in ($1, $2)`, tmpPending, tmpShipped)

	// Últimos 3 días y sólo pendientes: tiene que traer el propio y todo lo
	// que vuelva tiene que ser pendiente.
	orders, _, err := repo.listAll(ctx, StatusPending, now.AddDate(0, 0, -3), now.AddDate(0, 0, 1), 1, 100)
	if err != nil {
		t.Fatalf("listAll con filtros falló: %v", err)
	}
	var encontroPendiente bool
	for _, o := range orders {
		if o.Status != StatusPending {
			t.Errorf("esperaba sólo pendientes, vino %q", o.Status)
		}
		if o.Id == tmpPending {
			encontroPendiente = true
		}
	}
	if !encontroPendiente {
		t.Errorf("no apareció el pedido pendiente propio (%s)", tmpPending)
	}

	// Sólo por estado, sin fechas: el propio despachado y nada de otro estado.
	orders, _, err = repo.listAll(ctx, StatusShipped, nil, nil, 1, 100)
	if err != nil {
		t.Fatalf("listAll con status nulls falló: %v", err)
	}
	var encontroDespachado bool
	for _, o := range orders {
		if o.Status != StatusShipped {
			t.Errorf("esperaba sólo despachados, vino %q", o.Status)
		}
		if o.Id == tmpShipped {
			encontroDespachado = true
		}
	}
	if !encontroDespachado {
		t.Errorf("no apareció el pedido despachado propio (%s)", tmpShipped)
	}
}
