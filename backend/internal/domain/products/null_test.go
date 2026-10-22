package products

import (
	"context"
	"os"
	"testing"

	"example/internal/db"
)

// Lee una fila propia con portrait en null (los modelos usan puntero para esa
// columna: en string plano el scan revienta). La fila es del test y no de la
// semilla, porque el admin puede editar o crear platos desde la UI y cambiar
// los datos de la semilla. Corre contra la DB real; sin env, se saltea.
func TestReadsTolerateNullPortrait(t *testing.T) {
	if os.Getenv("POSTGRES_HOST") == "" {
		t.Skip("sin POSTGRES_* en el entorno")
	}
	ctx := context.Background()
	repo := NewRepo(db.GetConnection())

	// Resguardo por si un run anterior quedó a medias, y limpieza al final.
	if _, err := repo.db.Exec(ctx, `delete from products where name = 'zz-tmp-null-test'`); err != nil {
		t.Fatalf("limpieza previa: %v", err)
	}
	const tmpId = "bbbbbbbb-0000-4000-8000-000000000001"
	if _, err := repo.db.Exec(ctx,
		`insert into products (id, name, price) values ($1, 'zz-tmp-null-test', 100)`, tmpId); err != nil {
		t.Fatalf("insert de prueba: %v", err)
	}
	defer repo.db.Exec(ctx, `delete from products where id = $1`, tmpId)

	// findById: portrait null y category_id null (ambos anulables).
	tmp, err := repo.findById(ctx, tmpId)
	if err != nil {
		t.Fatalf("findById con portrait null falló: %v", err)
	}
	if tmp.Portrait != nil {
		t.Errorf("esperaba portrait null, obtuve %q", *tmp.Portrait)
	}
	if tmp.CategoryId != nil || tmp.CategoryName != nil {
		t.Errorf("esperaba sin categoría, obtuve %v / %v", tmp.CategoryId, tmp.CategoryName)
	}

	// list: el barrido con el join a categories no debe reventar.
	list, pages, err := repo.list(ctx, 1, 100, SORT_BY_NAME)
	if err != nil {
		t.Fatalf("list con join falló: %v", err)
	}
	if len(list) < 6 || pages < 1 {
		t.Fatalf("esperaba al menos los 6 platos de la semilla en %d página(s), obtuve %d", pages, len(list))
	}

	// listByName: la variante con ilike, contra la fila propia (determinístico).
	byName, _, err := repo.listByName(ctx, "zz-tmp-null", 1, 100, SORT_BY_PRICE)
	if err != nil {
		t.Fatalf("listByName con join falló: %v", err)
	}
	if len(byName) != 1 {
		t.Fatalf("esperaba 1 resultado para zz-tmp-null, obtuve %d", len(byName))
	}
}
