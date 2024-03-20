//go:build integration

package internal

import (
	"advertise_service/internal/handlers"
	"advertise_service/internal/infra"
	"advertise_service/internal/mock"
	"advertise_service/internal/models"
	"bytes"
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetAds(t *testing.T) {
	server := newProductionServer(infra.LoadConfig())
	requests := generatePostAdsRequests()
	for _, req := range requests {
		postAd(t, server, req)
	}
	getAds(t, server, "/api/v1/ad?limit=1000&offset=0&age=24&gender=F&country=TW&platform=ios")
	//get second time uses cache, so need additional testing
	getAds(t, server, "/api/v1/ad?limit=1000&offset=0&age=24&gender=M&country=JP&platform=web")
}

func getAds(t *testing.T, server http.Handler, url string) {
	request, err := http.NewRequest(http.MethodGet, url, nil)
	require.NoError(t, err)
	requestBody, err := handlers.ParseGetAdsRequest(request)
	conditionParam := handlers.ExtractConditionParams(requestBody)

	require.NoError(t, err)
	response := httptest.NewRecorder()
	server.ServeHTTP(response, request)
	assert.Equal(t, http.StatusOK, response.Code, response.Body.String())
	var resp handlers.GetAdsResponse
	err = json.Unmarshal(response.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.NotEmpty(t, resp.Items)

	for _, ad := range resp.Items {
		mock.MockedAdShouldShow(models.Ad{Title: ad.Title, EndAt: ad.EndAt}, conditionParam)
	}
}

func postAd(t *testing.T, server http.Handler, reqBody handlers.PostAdRequest) {
	jsonStr, err := json.Marshal(reqBody)
	require.NoError(t, err)
	request, err := http.NewRequest(http.MethodPost, "/api/v1/ad", bytes.NewBuffer(jsonStr))
	require.NoError(t, err)
	response := httptest.NewRecorder()
	server.ServeHTTP(response, request)
	assert.Equal(t, http.StatusCreated, response.Code, response.Body.String())
}

func generatePostAdsRequests() []handlers.PostAdRequest {
	var reqs []handlers.PostAdRequest
	for _, ad := range mock.GenerateMockAds() {
		reqs = append(reqs, handlers.PostAdRequest{
			Title:      ad.Title,
			StartAt:    ad.StartAt,
			EndAt:      ad.EndAt,
			Conditions: ad.Conditions,
		})
	}
	return reqs
}
