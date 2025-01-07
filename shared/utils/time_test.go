package utils_test

import (
	"testing"
	"time"

	"github.com/ryanjohnsontv/go-homeassistant/shared/utils"
	"github.com/stretchr/testify/assert"
)

func TestParseStateToTime(t *testing.T) {
	tests := []struct {
		name      string
		state     string
		expectErr bool
		validate  func(t *testing.T, result time.Time)
	}{
		{
			name:  "ISO8601 format",
			state: "2024-12-31T15:04:05Z",
			validate: func(t *testing.T, result time.Time) {
				assert.Equal(t, 2024, result.Year())
				assert.Equal(t, time.December, result.Month())
				assert.Equal(t, 31, result.Day())
				assert.Equal(t, 15, result.Hour())
				assert.Equal(t, 4, result.Minute())
				assert.Equal(t, 5, result.Second())
			},
		},
		{
			name:  "HH:MM:SS format",
			state: "12:34:56",
			validate: func(t *testing.T, result time.Time) {
				now := time.Now()
				assert.Equal(t, now.Year(), result.Year())
				assert.Equal(t, now.Month(), result.Month())
				assert.Equal(t, now.Day(), result.Day())
				assert.Equal(t, 12, result.Hour())
				assert.Equal(t, 34, result.Minute())
				assert.Equal(t, 56, result.Second())
			},
		},
		{
			name:  "Epoch format",
			state: "1704067200",
			validate: func(t *testing.T, result time.Time) {
				assert.Equal(t, 2023, result.Year())
				assert.Equal(t, time.December, result.Month())
				assert.Equal(t, 31, result.Day())
			},
		},
		{
			name:      "Invalid format",
			state:     "not-a-time",
			expectErr: true,
		},
		{
			name:      "Partial HH:MM format",
			state:     "12:34",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := utils.ParseStateToTime(tt.state)

			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.validate != nil {
					tt.validate(t, result)
				}
			}
		})
	}
}
