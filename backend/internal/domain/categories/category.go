package categories

import (
	"example/internal/domain/common"
)

// Categoría de platos: tabla maestra chica (nombre único, soft delete).
// La relación es products N:1 categories (products.category_id).
type Model struct {
	Id   string
	Name string
	common.TimeStamps
}
