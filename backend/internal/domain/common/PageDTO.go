package common

import (
	"net/url"
	"strconv"
)

type PageResponse[T any] struct {
	Data []T
	CurrentPage,
	TotalPages int
}

// PageParams lee ?page=&size= del querystring. Defaults: página 1 y 10
// elementos por página, con un tope de 100 para no pedir de más.
func PageParams(query url.Values) (page, size int) {
	if p, err := strconv.Atoi(query.Get("page")); err == nil && p > 0 {
		page = p
	} else {
		page = 1
	}
	if s, err := strconv.Atoi(query.Get("size")); err == nil && s > 0 {
		size = s
	} else {
		size = 10
	}
	if size > 100 {
		size = 100
	}
	return page, size
}

// TotalPagesFor calcula páginas totales redondeando hacia arriba.
func TotalPagesFor(totalRows, size int) int {
	if size < 1 {
		return 0
	}
	return (totalRows + size - 1) / size
}
