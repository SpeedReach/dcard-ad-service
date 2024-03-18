//go:build integration

package internal

import (
	"advertise_service/internal/handlers"
	"advertise_service/internal/infra"
	"advertise_service/internal/models"
	"bytes"
	"encoding/json"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestGetAds(t *testing.T) {
	server := NewServer(infra.LoadConfig())
	requests := generatePostAdsRequests()
	for _, req := range requests {
		postAd(t, server, req)
	}
	getAds(t, server)
}

func getAds(t *testing.T, server *http.ServeMux) {
	//get ads
	request, err := http.NewRequest(http.MethodGet, "/ad?limit=10&offset=0&age=24&gender=F&country=TW&platform=ios", nil)
	assert.NoError(t, err)
	response := httptest.NewRecorder()
	server.ServeHTTP(response, request)
	assert.Equal(t, http.StatusOK, response.Code)
	var resBody handlers.GetAdsResponse
	err = json.NewDecoder(response.Body).Decode(&resBody)
	assert.NoError(t, err)
	assert.NotEmpty(t, resBody.Items)

	params, _ := handlers.ParseGetAdsRequest(request)
	now := time.Now()
	var e time.Time
	for _, ad := range resBody.Items {
		assert.True(t, ad.EndAt.After(e))
		e = ad.EndAt
		assert.True(t, ad.EndAt.After(now))
		assert.True(t, shouldShow(ad.Title, params))
	}

}

func postAd(t *testing.T, server *http.ServeMux, reqBody handlers.PostAdRequest) {
	jsonStr, err := json.Marshal(reqBody)
	assert.NoError(t, err)
	request, err := http.NewRequest(http.MethodPost, "/ad", bytes.NewBuffer(jsonStr))
	assert.NoError(t, err)
	response := httptest.NewRecorder()
	server.ServeHTTP(response, request)
	assert.Equal(t, response.Code, http.StatusCreated, response.Body.String())
}

func shouldShow(adTitle string, queryParams handlers.GetAdsRequest) bool {
	if strings.Contains(adTitle, "inactive") {
		return false
	}

	//gender section
	if strings.Contains(adTitle, "F") || strings.Contains(adTitle, "M") {
		switch queryParams.Gender {
		case models.Female:
			return strings.Contains(adTitle, "F")
		case models.Male:
			return strings.Contains(adTitle, "M")
		}
	}

	//country section
	if strings.Contains(adTitle, "tw") || strings.Contains(adTitle, "jp") {
		switch queryParams.Country {
		case models.Taiwan:
			return strings.Contains(adTitle, "tw")
		case models.Japan:
			return strings.Contains(adTitle, "jp")
		}
		return false
	}

	//platform section
	if strings.Contains(adTitle, "web") || strings.Contains(adTitle, "ios") || strings.Contains(adTitle, "android") {
		switch queryParams.Platform {
		case models.Web:
			return strings.Contains(adTitle, "web")
		case models.Ios:
			return strings.Contains(adTitle, "ios")
		case models.Android:
			return strings.Contains(adTitle, "android")
		}
		return false
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

func generateTestAds() []models.Ad {
	return []models.Ad{
		{
			ID:      uuid.New(),
			StartAt: time.Now().Add(-time.Hour),
			EndAt:   time.Now().Add(time.Hour),
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
			StartAt: time.Now().Add(-time.Hour),
			EndAt:   time.Now().Add(time.Hour),
			Title:   "active",
		},
		{
			ID:      uuid.New(),
			StartAt: time.Now().Add(-time.Hour),
			EndAt:   time.Now().Add(time.Hour),
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
			StartAt: time.Now().Add(-time.Hour),
			EndAt:   time.Now().Add(time.Hour),
			Title:   "active_20_30_tw_jp_ios",
			Conditions: []models.Condition{
				{
					AgeStart: 20,
					AgeEnd:   30,
					Country: []models.Country{
						models.Taiwan,
						models.Japan,
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
}

func generatePostAdsRequests() []handlers.PostAdRequest {
	var reqs []handlers.PostAdRequest
	for _, ad := range generateTestAds() {
		reqs = append(reqs, handlers.PostAdRequest{
			Title:      ad.Title,
			StartAt:    ad.StartAt,
			EndAt:      ad.EndAt,
			Conditions: ad.Conditions,
		})
	}
	return reqs
}
