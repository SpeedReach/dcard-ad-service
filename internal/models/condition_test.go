package models

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCondition_Match(t *testing.T) {
	params1 := ConditionParams{
		Age:      20,
		Country:  Taiwan,
		Platform: Android,
		Gender:   Male,
	}
	params2 := ConditionParams{
		Age:      31,
		Country:  Japan,
		Platform: Ios,
		Gender:   Male,
	}

	cond1 := Condition{
		AgeStart: 18,
		AgeEnd:   30,
		Country:  []Country{Taiwan, Japan},
		Platform: []Platform{Android, Ios},
		Gender:   []Gender{Male},
	}
	cond2 := Condition{
		AgeStart: 18,
		AgeEnd:   40,
		Country:  []Country{Japan},
		Platform: []Platform{Ios},
		Gender:   []Gender{Male, Female},
	}
	assert.True(t, cond1.Match(params1))
	assert.False(t, cond1.Match(params2))
	assert.False(t, cond2.Match(params1))
	assert.True(t, cond2.Match(params2))

	cond3 := Condition{}

	assert.True(t, cond3.Match(params1))
	assert.True(t, cond3.Match(params2))
}
