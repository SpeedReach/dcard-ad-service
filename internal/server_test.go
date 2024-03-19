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
	getAds(t, server, "/api/v1/ad?limit=1000&offset=0&age=24&gender=F&country=TW&platform=ios")

	//get second time uses cache, so need additional testing
	getAds(t, server, "/api/v1/ad?limit=1000&offset=0&age=24&gender=M&country=JP&platform=web")
}

func getAds(t *testing.T, server *http.ServeMux, url string) {
	//get ads
	request, err := http.NewRequest(http.MethodGet, url, nil)
	assert.NoError(t, err)
	response := httptest.NewRecorder()
	server.ServeHTTP(response, request)
	assert.Equal(t, http.StatusOK, response.Code, response.Body.String())
	var resBody handlers.GetAdsResponse
	err = json.NewDecoder(response.Body).Decode(&resBody)
	assert.NoError(t, err)
	assert.NotEmpty(t, resBody.Items)

	params, _ := handlers.ParseGetAdsRequest(request)
	for _, ad := range resBody.Items {
		assert.True(t, shouldShow(ad.Title, params), ad.Title, url)
	}

}

func postAd(t *testing.T, server *http.ServeMux, reqBody handlers.PostAdRequest) {
	jsonStr, err := json.Marshal(reqBody)
	assert.NoError(t, err)
	request, err := http.NewRequest(http.MethodPost, "/api/v1/ad", bytes.NewBuffer(jsonStr))
	assert.NoError(t, err)
	response := httptest.NewRecorder()
	server.ServeHTTP(response, request)
	assert.Equal(t, http.StatusCreated, response.Code, response.Body.String())
}

func shouldShow(adTitle string, queryParams handlers.GetAdsRequest) bool {
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

func generateTestAds() []models.Ad {
	return []models.Ad{
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
