package rest

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ryanjohnsontv/go-homeassistant/shared/types"
	"github.com/stretchr/testify/assert"
)

func TestNewClient(t *testing.T) {
	t.Run("Valid Host and Token", func(t *testing.T) {
		client, err := NewClient("homeassistant.local", "test-token")
		assert.NoError(t, err)
		assert.NotNil(t, client)
		assert.Equal(t, "http://homeassistant.local:8123/api/", client.apiURL.String())
		assert.Equal(t, "Bearer test-token", client.bearerToken)
	})

	t.Run("Host With Scheme", func(t *testing.T) {
		client, err := NewClient("http://homeassistant.local", "test-token")
		assert.NoError(t, err)
		assert.NotNil(t, client)
		assert.Equal(t, "http://homeassistant.local:8123/api/", client.apiURL.String())
	})

	t.Run("Host With Port", func(t *testing.T) {
		client, err := NewClient("homeassistant.local:1234", "test-token")
		assert.NoError(t, err)
		assert.NotNil(t, client)
		assert.Equal(t, "http://homeassistant.local:1234/api/", client.apiURL.String())
	})

	t.Run("HTTPS Scheme", func(t *testing.T) {
		client, err := NewClient("homeassistant.local", "test-token", WithSecureConnection())
		assert.NoError(t, err)
		assert.NotNil(t, client)
		assert.Equal(t, "https://homeassistant.local:8123/api/", client.apiURL.String())
	})

	t.Run("Custom API Path", func(t *testing.T) {
		client, err := NewClient("homeassistant.local", "test-token", WithAPIPath("/custom/api/"))
		assert.NoError(t, err)
		assert.NotNil(t, client)
		assert.Equal(t, "http://homeassistant.local:8123/custom/api/", client.apiURL.String())
	})

	t.Run("Custom HTTP Client", func(t *testing.T) {
		customHTTPClient := &http.Client{}
		client, err := NewClient("homeassistant.local", "test-token", WithHTTPClient(customHTTPClient))
		assert.NoError(t, err)
		assert.NotNil(t, client)
		assert.Equal(t, customHTTPClient, client.httpClient)
	})

	t.Run("Empty Host", func(t *testing.T) {
		client, err := NewClient("", "test-token")
		assert.Error(t, err)
		assert.Nil(t, client)
	})

	t.Run("Empty Token", func(t *testing.T) {
		client, err := NewClient("homeassistant.local", "")
		assert.Error(t, err)
		assert.Nil(t, client)
	})
}

func TestClient(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message": "API running."}`))
	}))
	defer testServer.Close()

	client, err := NewClient(testServer.URL, "test-token")
	assert.NoError(t, err)

	t.Run("GetHealth", func(t *testing.T) {
		err := client.GetHealth()
		assert.NoError(t, err)
	})

	t.Run("GetHealth_Unhealthy", func(t *testing.T) {
		testServer.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"message": "Internal Server Error"}`))
		})

		err := client.GetHealth()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Internal Server Error")
	})

	t.Run("GetConfig", func(t *testing.T) {
		testServer.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			config := types.Config{
				Components: []string{"http", "websocket_api"},
				Latitude:   51.509865,
				Longitude:  -0.118092,
			}
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(config)
		})

		config, err := client.GetConfig()
		assert.NoError(t, err)
		assert.Contains(t, config.Components, "http")
		assert.Equal(t, 51.509865, config.Latitude)
		assert.Equal(t, -0.118092, config.Longitude)
	})

	t.Run("GetEvents", func(t *testing.T) {
		testServer.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			events := []Event{
				{Event: "state_changed", ListenerCount: 3},
				{Event: "time_changed", ListenerCount: 1},
			}
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(events)
		})

		events, err := client.GetEvents()
		assert.NoError(t, err)
		assert.Len(t, events, 2)
		assert.Equal(t, "state_changed", events[0].Event)
		assert.Equal(t, 3, events[0].ListenerCount)
	})

	t.Run("GetStates", func(t *testing.T) {
		testServer.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			states := []types.Entity{
				{
					EntityID: "light.kitchen",
					State:    "on",
				},
				{
					EntityID: "sensor.temperature",
					State:    "23.5",
				},
			}
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(states)
		})

		states, err := client.GetStates()
		assert.NoError(t, err)
		assert.Len(t, states, 2)
		assert.Equal(t, "light.kitchen", states[0].EntityID)
		assert.Equal(t, "on", states[0].State.String())
	})
}
