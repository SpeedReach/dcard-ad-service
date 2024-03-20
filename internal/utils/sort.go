package utils

import (
	"advertise_service/internal/models"
	"cmp"
	"slices"
)

func SortAdsByEndTimeAsc(ads []models.Ad) {
	slices.SortFunc(ads, func(a, b models.Ad) int {
		if a.EndAt.After(b.EndAt) {
			return 1
		} else {
			return -1
		}
	})
}

func SortAdsByID(ads []models.Ad) {
	slices.SortFunc(ads, func(a, b models.Ad) int {
		return cmp.Compare(a.ID.String(), b.ID.String())
	})

}
