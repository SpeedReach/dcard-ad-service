package models

import (
	"github.com/google/uuid"
	"time"
)

type Ad struct {
	ID         uuid.UUID   `json:"id"`
	Title      string      `json:"title"`
	StartAt    time.Time   `json:"start_at"`
	EndAt      time.Time   `json:"end_at"`
	Conditions []Condition `json:"conditions"`
}
