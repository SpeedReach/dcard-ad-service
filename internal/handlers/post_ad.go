package handlers

import (
	"advertise_service/internal/infra/cache"
	"advertise_service/internal/infra/logging"
	"advertise_service/internal/infra/persistent"
	"advertise_service/internal/models"
	"context"
	"encoding/json"
	"errors"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"log"
	"net/http"
	"time"
)

const MaxTitleLength = 100

type PostAdRequest struct {
	Title      string             `json:"title"`
	StartAt    time.Time          `json:"start_at"`
	EndAt      time.Time          `json:"end_at"`
	Conditions []models.Condition `json:"conditions"`
}

type PostAdResponse struct {
	//the created ad id
	AdID string
}

func PostAdHandler(writer http.ResponseWriter, request *http.Request) {
	//parse request
	reqBody := PostAdRequest{}
	err := json.NewDecoder(request.Body).Decode(&reqBody)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}

	//validate request
	err = validateRequest(reqBody)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}

	response, err := postAd(request.Context(), reqBody)

	if err != nil {
		http.Error(writer, "Internal error", http.StatusInternalServerError)
		return
	}

	//write response
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(writer).Encode(response)

	if err != nil {
		log.Printf("error encoding response: %v", err)
		http.Error(writer, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func postAd(ctx context.Context, reqBody PostAdRequest) (PostAdResponse, error) {
	logger := ctx.Value(logging.LoggerContextKey{}).(*zap.Logger)
	ad := models.Ad{
		ID:         uuid.New(),
		Title:      reqBody.Title,
		StartAt:    reqBody.StartAt,
		EndAt:      reqBody.EndAt,
		Conditions: reqBody.Conditions,
	}
	database := ctx.Value(StorageContextKey{}).(persistent.Storage)

	err := database.InsertAd(ctx, ad)
	if err != nil {
		return PostAdResponse{}, err
	}

	//store ad in cache if it's active the time that it's created
	if ad.StartAt.Before(time.Now()) {
		cacheService := ctx.Value(CacheContextKey{}).(cache.Service)
		err := cacheService.WriteActiveAd(ctx, ad)
		if err != nil {
			// It's ok that we failed to immediate cache the ad, scheduler will take care of it
			logger.Log(zap.ErrorLevel, "error caching active ad", zap.Error(err))
		}
	}

	return PostAdResponse{AdID: ad.ID.String()}, nil
}

func validateRequest(reqBody PostAdRequest) error {
	if reqBody.StartAt.After(reqBody.EndAt) {
		return errors.New("startAt must be before endAt")
	}
	if reqBody.EndAt.Before(time.Now()) {
		return errors.New("endAt must be in the future")
	}
	if len(reqBody.Title) > MaxTitleLength {
		return errors.New("title too long")
	}
	return nil
}
