package handlers

import (
	"advertise_service/internal/models"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func TestParseRequest(t *testing.T) {
	request, err := http.NewRequest("GET", "/ad?limit=3&offset=5&age=24&gender=F&country=TW&platform=ios", nil)
	assert.NoError(t, err)
	req, err := ParseGetAdsRequest(request)
	assert.NoError(t, err)
	assert.Equal(t, 3, req.Limit)
	assert.Equal(t, 5, req.Offset)
	assert.Equal(t, 24, req.Age)
	assert.Equal(t, models.Female, req.Gender)
	assert.Equal(t, models.Taiwan, req.Country)
	assert.Equal(t, models.Ios, req.Platform)
}
