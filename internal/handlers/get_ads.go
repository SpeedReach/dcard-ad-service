package handlers

import (
	"advertise_service/internal/infra"
	"advertise_service/internal/infra/cache"
	"advertise_service/internal/infra/logging"
	"advertise_service/internal/infra/persistent"
	"advertise_service/internal/models"
	"context"
	"encoding/json"
	"errors"
	"go.uber.org/zap"
	"log"
	"net/http"
	"strconv"
	"time"
)

type GetAdsRequest struct {
	Offset   int
	Limit    int
	Age      int
	Gender   models.Gender
	Platform models.Platform
	Country  models.Country
}

type GetAdsResponse struct {
	Items []item `json:"items"`
	//if there are no more active ads
	End bool `json:"end"`
}

type item struct {
	AdID  string    `json:"adId"`
	Title string    `json:"title"`
	EndAt time.Time `json:"endAt"`
}

func GetAdsHandler(writer http.ResponseWriter, request *http.Request) {
	reqParams, err := ParseGetAdsRequest(request)
	if err != nil {
		logging.ContextualLog(request.Context(), zap.ErrorLevel, "bad request", zap.Error(err))
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}
	conditionParams := extractConditionParams(reqParams)

	response, err := fetchMatched(request.Context(), reqParams, conditionParams)
	if err != nil {
		logging.ContextualLog(request.Context(), zap.ErrorLevel, "error fetching matched ads", zap.Error(err))
		http.Error(writer, "Internal Error", http.StatusInternalServerError)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(writer).Encode(response)
	if err != nil {
		log.Printf("error encoding response: %v", err)
		http.Error(writer, "Internal Error", http.StatusInternalServerError)
	}
}

// logic
func fetchMatched(ctx context.Context, reqParams GetAdsRequest, conditionParams models.ConditionParams) (GetAdsResponse, error) {
	cacheService := ctx.Value(infra.CacheContextKey{}).(cache.Service)
	db := ctx.Value(infra.DatabaseContextKey{}).(persistent.Database)

	activeAds, err := getActiveAdsCacheAside(ctx, cacheService, db, reqParams.Offset, reqParams.Limit)
	if err != nil {
		return GetAdsResponse{}, err
	}

	matched := 0
	matchedAds := make([]models.Ad, 0)
	for _, ad := range activeAds {
		if ad.ShouldShow(conditionParams) {
			matched++
			matchedAds = append(matchedAds, ad)
		}
	}

	response := GetAdsResponse{
		Items: make([]item, matched),
		End:   len(activeAds) < reqParams.Limit,
	}

	for i, ad := range matchedAds {
		response.Items[i] = item{
			AdID:  ad.ID.String(),
			Title: ad.Title,
			EndAt: ad.EndAt,
		}
	}

	return response, nil
}

// get active ads with cache aside method
func getActiveAdsCacheAside(ctx context.Context, cacheService cache.Service, db persistent.Database, skip int, count int) ([]models.Ad, error) {
	valid, err := cacheService.CheckCacheValid(ctx)
	if err != nil {
		return []models.Ad{}, err
	}

	if !valid {
		ads, err := db.FindAdsWithTime(ctx, time.Now().Add(80*time.Minute), time.Now())
		if err != nil {
			return []models.Ad{}, err
		}
		err = cacheService.WriteActiveAds(ctx, ads)
		return ads, err
	}

	return cacheService.GetActiveAds(ctx, skip, count)
}

// helper function for parsing request
func extractConditionParams(req GetAdsRequest) models.ConditionParams {
	return models.ConditionParams{
		Age:      req.Age,
		Gender:   req.Gender,
		Country:  req.Country,
		Platform: req.Platform,
	}
}

// ParseGetAdsRequest helper function for parsing request
func ParseGetAdsRequest(request *http.Request) (GetAdsRequest, error) {
	offsetStr := request.URL.Query().Get("offset")
	limitStr := request.URL.Query().Get("limit")
	ageStr := request.URL.Query().Get("age")
	gender := models.Gender(request.URL.Query().Get("gender"))
	country := models.Country(request.URL.Query().Get("country"))
	platform := models.Platform(request.URL.Query().Get("platform"))
	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 0 {
		limit = 5
	}
	age, err := strconv.Atoi(ageStr)
	if err != nil {
		return GetAdsRequest{}, err
	}
	if age < 0 {
		return GetAdsRequest{}, errors.New("age cannot be negative")
	}

	if !models.ValidCountry(country) {
		return GetAdsRequest{}, errors.New("invalid country")
	}
	if !models.ValidPlatform(platform) {
		return GetAdsRequest{}, errors.New("invalid platform")
	}
	if !models.ValidGender(gender) {
		return GetAdsRequest{}, errors.New("invalid gender")
	}

	parsed := GetAdsRequest{
		Offset:   offset,
		Limit:    limit,
		Age:      age,
		Gender:   gender,
		Country:  country,
		Platform: platform,
	}
	return parsed, nil
}
