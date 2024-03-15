package models

type Gender = string

const (
	Male   Gender = "M"
	Female Gender = "F"
)

func ValidGender(gender Gender) bool {
	switch gender {
	case Male, Female:
		return true
	default:
		return false
	}
}
