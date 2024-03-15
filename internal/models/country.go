package models

type Country string

const (
	Taiwan Country = "TW"
	Japan  Country = "JP"
)

func ValidCountry(country Country) bool {
	switch country {
	case Taiwan, Japan:
		return true
	}
	return false
}
