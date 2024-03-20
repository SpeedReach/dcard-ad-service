package handlers

import (
	"advertise_service/internal/infra/cache"
	"advertise_service/internal/infra/logging"
	"advertise_service/internal/infra/persistent"
	"advertise_service/internal/models"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"go.uber.org/zap"
	"net/http"
	"slices"
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
		logger.Log(zap.ErrorLevel, "bad request", zap.Error(err))
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}

	response, err := fetchMatched(request.Context(), reqParams)
	if err != nil {
		http.Error(writer, "Internal Error", http.StatusInternalServerError)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(writer).Encode(response)
	if err != nil {
		logger.Log(zap.ErrorLevel, "error encoding response", zap.Error(err))
		http.Error(writer, "Internal Error", http.StatusInternalServerError)
	}
}

// logic
func fetchMatched(ctx context.Context, reqParams GetAdsRequest) (GetAdsResponse, error) {
	logger := ctx.Value(logging.LoggerContextKey{}).(*zap.Logger)
	cacheService := ctx.Value(CacheContextKey{}).(cache.Service)
	db := ctx.Value(StorageContextKey{}).(persistent.Storage)

	activeAds, err := getActiveAdsCacheAside(ctx, cacheService, db, reqParams.Offset, reqParams.Limit)
	if err != nil {
		return GetAdsResponse{}, err
	}

	conditionParams := ExtractConditionParams(reqParams)
	matched := 0
	matchedAds := make([]models.Ad, 0)

	for _, ad := range activeAds {
		if ad.ShouldShow(conditionParams) {
			logger.Log(zap.DebugLevel, "ad matched", zap.String("ad", ad.String()), zap.String("params", conditionParams.String()))
			matched++
			matchedAds = append(matchedAds, ad)
		} else {
			logger.Log(zap.DebugLevel, "ad not matched", zap.String("ad", ad.String()), zap.String("params", fmt.Sprint(conditionParams)))
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
func getActiveAdsCacheAside(ctx context.Context, cacheService cache.Service, db persistent.Storage, skip int, count int) ([]models.Ad, error) {
	valid, err := cacheService.CheckCacheValid(ctx)
	if err != nil {
		logger.Log(zap.ErrorLevel, "error checking cache valid", zap.Error(err))
		return []models.Ad{}, err
	}

	now := time.Now().UTC()

	if !valid {
		logger.Log(zap.DebugLevel, "cache is invalid, fetching from database")
		ads, err := db.FindAdsWithTime(ctx, now.Add(80*time.Minute), now)
		if err != nil {
			logger.Log(zap.ErrorLevel, "error retrieving ads from database", zap.Error(err))
			return []models.Ad{}, err
		}
		err = cacheService.WriteActiveAds(ctx, ads)
		//since we fetch ads that will start in the future too, we need to filter them out before returning to the client
		ads = slices.DeleteFunc(ads, func(ad models.Ad) bool {
			return ad.StartAt.After(now)
		})
		return ads[min(skip, len(ads)):min(skip+count, len(ads))], err
	}

	logger.Log(zap.DebugLevel, "cache is valid, fetching from cache")
	return cacheService.GetActiveAds(ctx, skip, count)
}

// helper function for parsing request
func ExtractConditionParams(req GetAdsRequest) models.ConditionParams {
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
