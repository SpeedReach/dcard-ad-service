package database

import (
	"advertise_service/internal/models"
	"github.com/google/uuid"
	"slices"
)

type ConditionSchema struct {
	Id      uuid.UUID
	Ios     bool
	Android bool
	Web     bool
	Jp      bool
	Tw      bool
	Male    bool
	Female  bool
	MinAge  int
	MaxAge  int
}

func FromConditionModel(m models.Condition) ConditionSchema {
	ios := slices.Contains(m.Platform, models.Ios)
	android := slices.Contains(m.Platform, models.Android)
	web := slices.Contains(m.Platform, models.Web)
	jp := slices.Contains(m.Country, models.Japan)
	tw := slices.Contains(m.Country, models.Taiwan)
	male := slices.Contains(m.Gender, models.Male)
	female := slices.Contains(m.Gender, models.Female)
	return ConditionSchema{
		Id:      m.ConditionId,
		Ios:     ios,
		Android: android,
		Web:     web,
		Jp:      jp,
		Tw:      tw,
		MaxAge:  m.AgeEnd,
		MinAge:  m.AgeStart,
		Male:    male,
		Female:  female,
	}
}

func ToConditionModel(schema ConditionSchema) models.Condition {
	var platform []models.Platform
	var country []models.Country
	var genders []models.Gender
	if schema.Ios {
		platform = append(platform, models.Ios)
	}
	if schema.Android {
		platform = append(platform, models.Android)
	}
	if schema.Web {
		platform = append(platform, models.Web)
	}
	if schema.Jp {
		country = append(country, models.Japan)
	}
	if schema.Tw {
		country = append(country, models.Taiwan)
	}
	if schema.Male {
		genders = append(genders, models.Male)
	}
	if schema.Female {
		genders = append(genders, models.Female)
	}

	return models.Condition{
		ConditionId: schema.Id,
		Platform:    platform,
		Country:     country,
		Gender:      genders,
		AgeEnd:      schema.MaxAge,
		AgeStart:    schema.MinAge,
	}
}
