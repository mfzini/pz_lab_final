package common

type PageResponse[T any] struct {
	Data []T
	CurrentPage,
	TotalPages int
}
