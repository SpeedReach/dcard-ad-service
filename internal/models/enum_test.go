package models

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestValidEnum(t *testing.T) {
	assert.True(t, ValidCountry("TW"))
	assert.True(t, ValidCountry("JP"))
	assert.False(t, ValidCountry("US"))
	assert.False(t, ValidCountry("CN"))
	assert.False(t, ValidCountry("KR"))
	assert.True(t, ValidPlatform("android"))
	assert.True(t, ValidPlatform("ios"))
	assert.False(t, ValidPlatform("windows"))
	assert.False(t, ValidPlatform("Mac"))
	assert.False(t, ValidPlatform("Linux"))
	assert.True(t, ValidGender("M"))
	assert.True(t, ValidGender("F"))
	assert.False(t, ValidGender("X"))
	assert.False(t, ValidGender("Y"))

}
