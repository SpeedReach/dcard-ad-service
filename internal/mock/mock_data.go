//go:build test

package mock

import (
	"advertise_service/internal/models"
	"github.com/google/uuid"
	"strconv"
	"strings"
	"time"
)

// MockedAdShouldShow determine if a mocked ad should be shown
func MockedAdShouldShow(ad models.Ad, queryParams models.ConditionParams) bool {
	if ad.StartAt.After(time.Now().UTC()) || ad.EndAt.Before(time.Now().UTC()) {
		return false
	}

	adTitle := ad.Title
	if strings.Contains(adTitle, "inactive") {
		return false
	}

	//gender
	if strings.Contains(adTitle, "F") && !strings.Contains(adTitle, "M") {
		if queryParams.Gender == models.Male {
			return false
		}
	}
	if strings.Contains(adTitle, "M") && !strings.Contains(adTitle, "F") {
		if queryParams.Gender == models.Female {
			return false
		}
	}

	//country section
	if strings.Contains(adTitle, "tw") && !strings.Contains(adTitle, "jp") {
		if queryParams.Country != models.Taiwan {
			return false
		}
	}
	if strings.Contains(adTitle, "jp") && !strings.Contains(adTitle, "tw") {
		if queryParams.Country != models.Japan {
			return false
		}
	}

	//platform section
	if strings.Contains(adTitle, "web") || strings.Contains(adTitle, "ios") || strings.Contains(adTitle, "android") {
		switch queryParams.Platform {
		case models.Web:
			if !strings.Contains(adTitle, "web") {
				return false
			}
		case models.Ios:
			if !strings.Contains(adTitle, "ios") {
				return false
			}
		case models.Android:
			if !strings.Contains(adTitle, "android") {
				return false
			}
		}
	}

	//age section
	args := strings.Split(adTitle, "_")
	if len(args) >= 2 {
		ageStart, err := strconv.Atoi(args[1])
		if err != nil {
			return true
		}
		ageEnd, err := strconv.Atoi(args[2])
		if err != nil {
			return true
		}
		return ageStart <= queryParams.Age && queryParams.Age <= ageEnd
	}

	return true
}

func GenerateMockAds() []models.Ad {
	ads := []models.Ad{
		{
			ID:      uuid.New(),
			StartAt: time.Now().UTC().Add(-time.Hour),
			EndAt:   time.Now().UTC().Add(time.Hour),
			Title:   "active_20_30_tw_web",
			Conditions: []models.Condition{
				{
					AgeStart: 20,
					AgeEnd:   30,
					Country: []models.Country{
						models.Taiwan,
					},
					Platform: []models.Platform{
						models.Web,
					},
				},
			},
		},
		{
			ID:      uuid.New(),
			StartAt: time.Now().UTC().Add(-time.Hour),
			EndAt:   time.Now().UTC().Add(time.Hour),
			Title:   "active",
		},
		{
			ID:      uuid.New(),
			StartAt: time.Now().UTC().Add(-time.Hour),
			EndAt:   time.Now().UTC().Add(time.Hour),
			Title:   "active_50_60_jp_ios_android_M",
			Conditions: []models.Condition{
				{
					AgeStart: 50,
					AgeEnd:   60,
					Country: []models.Country{
						models.Japan,
					},
					Platform: []models.Platform{
						models.Ios,
						models.Android,
					},
					Gender: []models.Gender{
						models.Male,
					},
				},
			},
		},
		{
			ID:      uuid.New(),
			StartAt: time.Now().UTC().Add(-time.Hour),
			EndAt:   time.Now().UTC().Add(time.Hour),
			Title:   "active_20_30_tw_jp_ios",
			Conditions: []models.Condition{
				{
					AgeStart: 20,
					AgeEnd:   30,
					Country: []models.Country{
						models.Taiwan,
						models.Japan,
					},
					Platform: []models.Platform{
						models.Ios,
					},
				},
			},
		},
		{
			ID:      uuid.New(),
			StartAt: time.Now().UTC().Add(1000 * time.Hour),
			EndAt:   time.Now().UTC().Add(1001 * time.Hour),
			Title:   "inactive",
		},
		{
			ID:      uuid.New(),
			StartAt: time.Now().UTC().Add(-time.Hour),
			EndAt:   time.Now().UTC().Add(time.Hour),
			Title:   "active",
		},
		{
			ID:      uuid.New(),
			StartAt: time.Now().UTC().Add(-time.Hour),
			EndAt:   time.Now().UTC().Add(time.Hour),
			Title:   "active_20_30_tw_web_F",
			Conditions: []models.Condition{
				{
					AgeStart: 20,
					AgeEnd:   30,
					Country: []models.Country{
						models.Taiwan,
					},
					Platform: []models.Platform{
						models.Web,
					},
					Gender: []models.Gender{
						models.Female,
					},
				},
			},
		},
	}
	return ads
}
