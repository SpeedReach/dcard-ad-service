package models

import (
	"github.com/google/uuid"
	"slices"
)

type Condition struct {
	ConditionId uuid.UUID  `json:"conditionId"`
	AgeStart    int        `json:"ageStart"`
	AgeEnd      int        `json:"ageEnd"`
	Country     []Country  `json:"country"`
	Gender      []Gender   `json:"gender"`
	Platform    []Platform `json:"platform"`
}

type ConditionParams struct {
	Age      int      `json:"age"`
	Gender   Gender   `json:"gender"`
	Country  Country  `json:"country"`
	Platform Platform `json:"platform"`
}

func (c Condition) Match(p ConditionParams) bool {
	if !((p.Age > c.AgeStart && p.Age < c.AgeEnd) || (c.AgeStart == 0 && c.AgeEnd == 0)) {
		return false
	}
	if len(c.Platform) > 0 {
		match := slices.Contains(c.Platform, p.Platform)
		if !match {
			return false
		}
	}

	if len(c.Country) > 0 {
		match := slices.Contains(c.Country, p.Country)
		if !match {
			return false
		}
	}

	if len(c.Gender) > 0 {
		match := slices.Contains(c.Gender, p.Gender)
		if !match {
			return false
		}
	}
	return true
}
