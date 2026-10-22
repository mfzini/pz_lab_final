package users

import (
	"os"
	"testing"

	"example/internal/db"
)

// María se registra sin teléfono: phone es anulable y tiene que leerse como
// null (puntero), no como string plano. El admin sirve de control positivo
// (su teléfono existe).
func TestReadsTolerateNullPhone(t *testing.T) {
	if os.Getenv("POSTGRES_HOST") == "" {
		t.Skip("sin POSTGRES_* en el entorno")
	}
	repo := NewRepo(db.GetConnection())

	maria, err := repo.FindById("00000000-0000-4000-8000-000000000003")
	if err != nil {
		t.Fatalf("FindById con phone null falló: %v", err)
	}
	if maria.Phone != nil {
		t.Errorf("esperaba phone null, obtuve %q", *maria.Phone)
	}

	if _, err := repo.FindByEmail("maria@test.com"); err != nil {
		t.Fatalf("FindByEmail con phone null falló: %v", err)
	}

	admin, err := repo.FindById("00000000-0000-4000-8000-000000000001")
	if err != nil {
		t.Fatalf("FindById del admin falló: %v", err)
	}
	if admin.Phone == nil {
		t.Errorf("el admin sí tiene teléfono: esperaba uno")
	}
}
