package utils

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

func ParseStateToTime(state string) (time.Time, error) {
	var t time.Time

	var err error

	// ISO8601 or Full Timestamp
	if strings.Contains(state, "T") {
		t, err = time.Parse(time.RFC3339, state)
		if err == nil {
			return t, nil
		}
	}

	// HH:MM:SS Format
	if strings.Contains(state, ":") {
		now := time.Now()
		parts := strings.Split(state, ":")

		if len(parts) == 3 {
			hour, _ := strconv.Atoi(parts[0])
			minute, _ := strconv.Atoi(parts[1])
			second, _ := strconv.Atoi(parts[2])

			t = time.Date(now.Year(), now.Month(), now.Day(), hour, minute, second, 0, now.Location())

			return t, nil
		}
	}

	// Epoch Format
	epoch, err := strconv.ParseInt(state, 10, 64)
	if err == nil {
		return time.Unix(epoch, 0), nil
	}

	return t, fmt.Errorf("unsupported state format: %s", state)
}
