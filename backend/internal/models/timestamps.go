package models

import "time"

type TimeStamps struct {
	CreatedAt,
	UpdatedAt,
	DeletedAt *time.Time
}
