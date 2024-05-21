package homeassistant

import (
	"time"
)

// TriggerDateTime initiates a ticker that triggers callbacks at specified times.
func (c *Client) TriggerDateTime() {
	if len(c.dateTimeEntityListeners) == 0 {
		return
	}
	// Wait until the start of the next minute to begin the ticker.
	now := time.Now()
	untilNextMinute := time.Until(now.Truncate(time.Minute).Add(time.Minute))
	time.Sleep(untilNextMinute)

	// Set up a ticker that ticks every minute.
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for t := range ticker.C {
		c.triggerCallbacks(t)
	}
}

// triggerCallbacks triggers the appropriate callbacks for a given time.
func (c *Client) triggerCallbacks(t time.Time) {
	currentHour, currentMinute, _ := t.Clock()
	for entityTime, dateTimeMap := range c.dateTimeEntityListeners {
		entityHour, entityMinute, _ := entityTime.Clock()

		// Check if the current hour and minute match any scheduled times.
		if currentHour == entityHour && currentMinute == entityMinute {
			c.executeCallbacks(dateTimeMap)
		}
	}
}

// executeCallbacks executes all callbacks associated with a specific time.
func (c *Client) executeCallbacks(dateTimeMap map[string][]dateTimeEntityTrigger) {
	for entityID, functions := range dateTimeMap {
		for _, eventTrigger := range functions {
			c.logger.Debug().
				Str("entity_id", entityID).
				Msg("Triggering datetime entity callback function")
			if eventTrigger.timeType == "time" {
				go eventTrigger.callback()
			}
		}
	}
}
