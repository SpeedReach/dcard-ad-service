package models

import (
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestShouldShow(t *testing.T) {
	ads := generateTestAds()
	params1 := ConditionParams{
		Age:      25,
		Country:  Taiwan,
		Platform: Web,
		Gender:   Male,
	}
	assert.True(t, ads[0].ShouldShow(params1))
	assert.True(t, ads[1].ShouldShow(params1))
	assert.False(t, ads[2].ShouldShow(params1))
	assert.False(t, ads[3].ShouldShow(params1))
	assert.False(t, ads[4].ShouldShow(params1))
	assert.True(t, ads[5].ShouldShow(params1))
	assert.False(t, ads[6].ShouldShow(params1))
}

func generateTestAds() []Ad {
	return []Ad{
		{
			ID:      uuid.New(),
			StartAt: time.Now().Add(-time.Hour),
			EndAt:   time.Now().Add(time.Hour),
			Title:   "active_20_30_tw_web",
			Conditions: []Condition{
				{
					AgeStart: 20,
					AgeEnd:   30,
					Country: []Country{
						Taiwan,
					},
					Platform: []Platform{
						Web,
					},
				},
			},
		},
		{
			ID:      uuid.New(),
			StartAt: time.Now().Add(-time.Hour),
			EndAt:   time.Now().Add(time.Hour),
			Title:   "active",
		},
		{
			ID:      uuid.New(),
			StartAt: time.Now().Add(-time.Hour),
			EndAt:   time.Now().Add(time.Hour),
			Title:   "active_50_60_jp_ios_android_M",
			Conditions: []Condition{
				{
					AgeStart: 50,
					AgeEnd:   60,
					Country: []Country{
						Japan,
					},
					Platform: []Platform{
						Ios,
						Android,
					},
					Gender: []Gender{
						Male,
					},
				},
			},
		},
		{
			ID:      uuid.New(),
			StartAt: time.Now().Add(-time.Hour),
			EndAt:   time.Now().Add(time.Hour),
			Title:   "active_20_30_tw_jp_ios",
			Conditions: []Condition{
				{
					AgeStart: 20,
					AgeEnd:   30,
					Country: []Country{
						Taiwan,
						Japan,
					},
					Platform: []Platform{
						Ios,
					},
				},
			},
		},
		{
			ID:      uuid.New(),
			StartAt: time.Now().Add(1000 * time.Hour),
			EndAt:   time.Now().Add(1001 * time.Hour),
			Title:   "inactive",
		},
		{
			ID:      uuid.New(),
			StartAt: time.Now().Add(-time.Hour),
			EndAt:   time.Now().Add(time.Hour),
			Title:   "active",
		},
		{
			ID:      uuid.New(),
			StartAt: time.Now().Add(-time.Hour),
			EndAt:   time.Now().Add(time.Hour),
			Title:   "active_20_30_tw_web_F",
			Conditions: []Condition{
				{
					AgeStart: 20,
					AgeEnd:   30,
					Country: []Country{
						Taiwan,
					},
					Platform: []Platform{
						Web,
					},
					Gender: []Gender{
						Female,
					},
				},
			},
		},
	}
}
