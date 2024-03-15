package internal

import (
	"advertise_service/internal/handlers"
	"advertise_service/internal/infra"
	"advertise_service/internal/models"
	"bytes"
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestGetAds(t *testing.T) {
	server := NewServer(infra.LoadConfig())
	reqBody := handlers.PostAdRequest{
		Title:   "testTitle",
		StartAt: time.Now().Add(-1 * time.Minute),
		EndAt:   time.Now().Add(time.Minute),
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
	}
	postAd(t, server, reqBody)

}

func getAds(t *testing.T, server *http.ServeMux) {
	request, err := http.NewRequest(http.MethodGet, "/ad", nil)
	assert.NoError(t, err)
	response := httptest.NewRecorder()
	server.ServeHTTP(response, request)
	assert.Equal(t, response.Code, http.StatusOK)
}

func postAd(t *testing.T, server *http.ServeMux, reqBody handlers.PostAdRequest) {

	jsonStr, err := json.Marshal(reqBody)
	assert.NoError(t, err)
	request, err := http.NewRequest(http.MethodPost, "/ad", bytes.NewBuffer(jsonStr))
	assert.NoError(t, err)
	response := httptest.NewRecorder()
	server.ServeHTTP(response, request)
	assert.Equal(t, response.Code, http.StatusCreated)
}
