package models

import (
	"github.com/google/uuid"
	"time"
)

type NewsPreview struct {
	Uuid      uuid.UUID `db:"uuid"`
	Title     string    `db:"title"`
	Link      string    `db:"link"`
	Content   *string   `db:"content"`
	Status    string    `db:"status"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}
