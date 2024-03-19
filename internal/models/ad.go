package models

import (
	"encoding/json"
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

func (ad Ad) ShouldShow(params ConditionParams) bool {
	if !ad.IsActive() {
		return false
	}
	if len(ad.Conditions) == 0 {
		return true
	}
	for _, condition := range ad.Conditions {
		if condition.Match(params) {
			return true
		}
	}
	return false
}

func (ad Ad) IsActive() bool {
	now := time.Now().UTC()
	return now.After(ad.StartAt) && now.Before(ad.EndAt)
}

func (ad Ad) String() string {
	jStr, _ := json.Marshal(ad)
	return string(jStr)
}
