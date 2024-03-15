package handlers

import (
	"advertise_service/internal/cache"
	"advertise_service/internal/models"
	"context"
	"encoding/json"
	"errors"
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
	Cursor int    `json:"cursor"`
	Items  []item `json:"items"`
}

type item struct {
	AdID  string    `json:"adId"`
	Title string    `json:"title"`
	EndAt time.Time `json:"endAt"`
}

func GetAds(writer http.ResponseWriter, request *http.Request) {
	reqParams, err := parseGetAds(request)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}
	conditionParams := extractConditionParams(reqParams)

	response, err := fetchMatched(request.Context(), reqParams, conditionParams)
	if err != nil {
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

// continuously fetch matched active ads from cache until we have enough or there are no more ads
func fetchMatched(ctx context.Context, reqParams GetAdsRequest, conditionParams models.ConditionParams) (GetAdsResponse, error) {
	matchedAds := make([]models.Ad, reqParams.Limit)
	matched := 0
	cursor := reqParams.Offset
	for matched < reqParams.Limit {
		count := reqParams.Limit - matched
		ads, err := cache.GetActiveAds(ctx, cursor, count)
		cursor += len(ads)

		if err != nil {
			log.Printf("error getting ads: %v", err)
			return GetAdsResponse{}, err
		}
		for _, ad := range ads {
			for _, condition := range ad.Conditions {
				if condition.Match(conditionParams) {
					matchedAds[matched] = ad
					matched++
					break
				}
			}
		}

		if len(ads) < count {
			break
		}
	}

	response := GetAdsResponse{
		Cursor: cursor,
		Items:  make([]item, matched),
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

// helper function for parsing request
func extractConditionParams(req GetAdsRequest) models.ConditionParams {
	return models.ConditionParams{
		Age:      req.Age,
		Gender:   req.Gender,
		Country:  req.Country,
		Platform: req.Platform,
	}
}

// helper function for parsing request
func parseGetAds(request *http.Request) (GetAdsRequest, error) {
	offsetStr := request.URL.Query().Get("offset")
	limitStr := request.URL.Query().Get("limit")
	ageStr := request.URL.Query().Get("age")
	gender := models.Gender(request.URL.Query().Get("gender"))
	country := models.Country(request.URL.Query().Get("country"))
	platform := models.Platform(request.URL.Query().Get("platform"))
	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		return GetAdsRequest{}, err
	}
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		return GetAdsRequest{}, err
	}
	age, err := strconv.Atoi(ageStr)
	if err != nil {
		return GetAdsRequest{}, err
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
