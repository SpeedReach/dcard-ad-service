package handlers

import (
	"advertise_service/internal/cache"
	"advertise_service/internal/database"
	"advertise_service/internal/models"
	"encoding/json"
	"errors"
	"github.com/google/uuid"
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

func PostAd(writer http.ResponseWriter, request *http.Request) {
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

	//insert db
	ad := models.Ad{
		ID:         uuid.New(),
		Title:      reqBody.Title,
		StartAt:    reqBody.StartAt,
		EndAt:      reqBody.EndAt,
		Conditions: reqBody.Conditions,
	}
	err = database.InsertAd(request.Context(), ad)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}

	//store ad in cache if it's active the time that it's created
	if ad.StartAt.Before(time.Now()) {
		err := cache.StoreActiveAd(request.Context(), ad)
		if err != nil {
			// It's ok that we failed to immediate cache the ad, scheduler will take care of it
			log.Printf("error caching active ad: %v", err)
		}
	}

	//write response
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(writer).Encode(PostAdResponse{AdID: ad.ID.String()})

	if err != nil {
		log.Printf("error encoding response: %v", err)
		http.Error(writer, "Internal Server Error", http.StatusInternalServerError)
		return
	}
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
