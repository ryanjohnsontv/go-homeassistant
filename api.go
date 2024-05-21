// A Go client for communicating with Home Assistant's REST API.
// https://developers.home-assistant.io/docs/api/rest

package homeassistant

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

func (c *Client) newGETRequest(path string) (*http.Request, error) {
	return http.NewRequest(http.MethodGet, c.apiURL+path, nil)
}

func (c *Client) newPOSTRequest(path string, payload interface{}) (*http.Request, error) {
	b, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	return http.NewRequest(http.MethodPost, c.apiURL+path, bytes.NewBuffer(b))
}

// SendRequest sends an HTTP request and returns the response.
// Pass through pointer to decode a JSON response.
func (c *Client) sendRequest(req *http.Request, v any) error {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", c.accessToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.logger.Error().
			Msgf("unexpected response code: %d", resp.StatusCode)
		return ErrUnhealthyAPI
	}

	if v != nil {
		if err := json.NewDecoder(resp.Body).Decode(&v); err != nil {
			return err
		}
	}

	return nil
}

// GetHealth returns a message if the API is up and running.
func (c *Client) APIGetHealth() (string, error) {
	req, err := c.newGETRequest("")
	if err != nil {
		return "", err
	}

	var message map[string]string

	if err = c.sendRequest(req, &message); err != nil {
		return "", err
	}

	return message["message"], nil
}

// GetConfig gets the Home Assistant configuration.
func (c *Client) APIGetConfig() (HomeAssistantConfig, error) {
	req, err := c.newGETRequest("config")
	if err != nil {
		return HomeAssistantConfig{}, err
	}

	var resp HomeAssistantConfig

	if err = c.sendRequest(req, &resp); err != nil {
		return HomeAssistantConfig{}, err
	}

	return resp, nil
}

type event struct {
	Event         string `json:"event"`
	ListenerCount int    `json:"listener_count"`
}

// GetEvents gets a list of all events in Home Assistant.
func (c *Client) APIGetEvents() ([]event, error) {
	req, err := c.newGETRequest("events")
	if err != nil {
		return nil, err
	}

	var events []event

	if err = c.sendRequest(req, &events); err != nil {
		return nil, err
	}

	return events, nil
}

type Service struct {
	Domain   string   `json:"domain"`
	Services []string `json:"services"`
}

// GetServices gets a list of all services in Home Assistant.
func (c *Client) APIGetServices() ([]Service, error) {
	req, err := c.newGETRequest("services")
	if err != nil {
		return nil, err
	}

	var services []Service

	if err = c.sendRequest(req, &services); err != nil {
		return nil, err
	}

	return services, nil
}

// GetHistory gets the history of events and state changes in Home Assistant.
// func (c *Client) APIGetHistory(start time.Time, end time.Time, filter map[string]string) ([]map[string]interface{}, error) {
// 	url := fmt.Sprintf("%s/api/history/period/%s", c.apiURL, timestamp)
// 	queryParams := make(url.Values)
// 	queryParams.Add("start", start.Format(time.RFC3339))
// 	queryParams.Add("end", end.Format(time.RFC3339))
// 	for key, value := range filter {
// 		queryParams.Add(key, value)
// 	}
// 	req, err := c.newGETRequest(fmt.Sprintf("%s?%s", url, queryParams.Encode()))
// 	if err != nil {
// 		return nil, err
// 	}

// 	resp, err := c.httpClient.Do(req)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer resp.Body.Close()

// 	if resp.StatusCode != http.StatusOK {
// 		return nil, fmt.Errorf("failed to get history")
// 	}

// 	var history []map[string]interface{}
// 	if err := json.NewDecoder(resp.Body).Decode(&history); err != nil {
// 		return nil, err
// 	}

// 	return history, nil
// }

// GetLogbook gets the logbook of events in Home Assistant.
// func (c *Client) APIGetLogbook(start time.Time, end time.Time, filter map[string]string) ([]map[string]interface{}, error) {
// 	url := fmt.Sprintf("%s/api/logbook", c.apiURL)
// 	queryParams := make(url.Values)
// 	queryParams.Add("start", start.Format(time.RFC3339))
// 	queryParams.Add("end", end.Format(time.RFC3339))
// 	for key, value := range filter {
// 		queryParams.Add(key, value)
// 	}
// 	req, err := c.newGETRequest(fmt.Sprintf("%s?%s", url, queryParams.Encode()))
// 	if err != nil {
// 		return nil, err
// 	}

// 	client := &http.Client{}
// 	resp, err := client.Do(req)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer resp.Body.Close()

// 	if resp.StatusCode != http.StatusOK {
// 		return nil, fmt.Errorf("failed to get logbook")
// 	}

// 	var logbook []map[string]interface{}
// 	if err := json.NewDecoder(resp.Body).Decode(&logbook); err != nil {
// 		return nil, err
// 	}

// 	return logbook, nil
// }

// GetStates gets a list of all states in Home Assistant.
func (c *Client) APIGetStates() ([]State, error) {
	req, err := c.newGETRequest("states")
	if err != nil {
		return nil, err
	}

	var resp []State

	if err = c.sendRequest(req, &resp); err != nil {
		return nil, err
	}

	return resp, nil
}

// GetState gets the state of an entity in Home Assistant.
func (c *Client) APIGetState(entityID string) (State, error) {
	req, err := c.newGETRequest(fmt.Sprintf("states/%s", entityID))
	if err != nil {
		return State{}, err
	}

	var resp State

	if err = c.sendRequest(req, &resp); err != nil {
		return State{}, err
	}

	return resp, nil
}

// GetErrorLog gets the error log in Home Assistant.
func (c *Client) APIGetErrorLog() ([]map[string]interface{}, error) {
	req, err := c.newGETRequest("error_log")
	if err != nil {
		return nil, err
	}

	var errorLog []map[string]interface{}

	if err = c.sendRequest(req, &errorLog); err != nil {
		return nil, err
	}

	return errorLog, nil
}

// GetCameraProxy gets a proxy URL for a camera in Home Assistant.
func (c *Client) APIGetCameraProxy(entityID string) (string, error) {
	req, err := c.newGETRequest(fmt.Sprintf("camera_proxy/%s", entityID))
	if err != nil {
		return "", err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to get camera proxy")
	}

	return resp.Header.Get("Location"), nil
}

// GetCalendars gets a list of calendar entities in Home Assistant.
func (c *Client) APIGetCalendars(calendarID string) (map[string]interface{}, error) {
	req, err := c.newGETRequest(fmt.Sprintf("calendars/%s", calendarID))
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get calendar")
	}

	var calendar map[string]interface{}

	if err := json.NewDecoder(resp.Body).Decode(&calendar); err != nil {
		return nil, err
	}

	return calendar, nil
}

type calendarEvent struct {
	Summary     string       `json:"summary"`
	Start       calendarDate `json:"start"`
	End         calendarDate `json:"end"`
	Description string       `json:"description"`
	Location    string       `json:"location"`
}

type calendarDate struct {
	Date     string    `json:"date"`
	DateTime time.Time `json:"dateTime"`
}

type CalendarEvents []calendarEvent

// GetCalendarEvents gets the events of a calendar in Home Assistant.
func (c *Client) APIGetCalendarEvents(calendarID string, start *time.Time, end *time.Time) (CalendarEvents, error) {
	req, err := c.newGETRequest(
		fmt.Sprintf("%scalendars/%s?start=%s&end=",
			calendarID,
			start.Format(time.RFC3339),
			end.Format(time.RFC3339)))
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()

	if start != nil {
		q.Add("start", start.Format(time.RFC3339))
	}

	if end != nil {
		q.Add("end", end.Format(time.RFC3339))
	}

	req.URL.RawQuery = q.Encode()

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get calendar events")
	}

	var events CalendarEvents

	if err := json.NewDecoder(resp.Body).Decode(&events); err != nil {
		return nil, err
	}

	return events, nil
}

// UpsertState updates or creates a state in Home Assistant.
// Returns a state object and a URL of the new resource if one is created.
func (c *Client) APIUpsertState(entityID string, entityState map[string]interface{}) (State, *url.URL, error) {
	req, err := c.newPOSTRequest(fmt.Sprintf("states/%s", entityID), entityState)
	if err != nil {
		return State{}, nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return State{}, nil, err
	}

	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		var v State
		if err := json.NewDecoder(resp.Body).Decode(&v); err != nil {
			return State{}, nil, err
		}

		return v, nil, err
	case http.StatusCreated:
		var v State
		if err := json.NewDecoder(resp.Body).Decode(&v); err != nil {
			return State{}, nil, err
		}

		newResource, err := url.Parse(resp.Header.Get("Location"))

		return v, newResource, err
	default:
		return State{}, nil, fmt.Errorf("failed to upsert state")
	}
}

// APIFireEvent fires an event in Home Assistant.
// Returns a message if successful.
func (c *Client) APIFireEvent(eventType string, eventData map[string]interface{}) (string, error) {
	req, err := c.newPOSTRequest(fmt.Sprintf("events/%s", eventType), eventData)
	if err != nil {
		return "", err
	}

	var resp map[string]string

	if err = c.sendRequest(req, &resp); err != nil {
		return "", err
	}

	return resp["message"], nil
}

// APICallService calls a Home Assistant service via the REST API.
// Returns a list of states that have changed while the service was being executed.
func (c *Client) APICallService(domain string, service string, data map[string]interface{}) ([]State, error) {
	req, err := c.newPOSTRequest(fmt.Sprintf("services/%s/%s", domain, service), data)
	if err != nil {
		return nil, err
	}

	var resp []State

	if err = c.sendRequest(req, &resp); err != nil {
		return nil, err
	}

	return resp, nil
}

// RenderTemplate renders a Home Assistant template.
func (c *Client) APIRenderTemplate(template string, variables map[string]interface{}) (string, error) {
	data := map[string]interface{}{
		"template":  template,
		"variables": variables,
	}

	req, err := c.newPOSTRequest("template", data)
	if err != nil {
		return "", err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to render template")
	}

	responseData, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(responseData), nil
}

type checkConfig struct {
	Errors *string `json:"errors"`
	Result string  `json:"result"`
}

// CheckConfig checks the Home Assistant configuration.
// If the checks is successful, nil will be returned.
// If the check fails, a string containing the error will be returned.
func (c *Client) APICheckConfig() (*string, error) {
	req, err := c.newPOSTRequest("config/core/check_config", nil)
	if err != nil {
		return nil, err
	}

	var resp checkConfig

	if err = c.sendRequest(req, &resp); err != nil {
		return nil, err
	}

	if resp.Result == "valid" {
		return nil, nil
	}

	return resp.Errors, nil
}

// HandleIntent handles an intent in Home Assistant.
// You must add intent: to your Home Assistant configuration file to enable this endpoint.
func (c *Client) APIHandleIntent(intent map[string]interface{}) error {
	req, err := c.newPOSTRequest("services/intent/handle", intent)
	if err != nil {
		return err
	}

	var resp checkConfig

	if err = c.sendRequest(req, &resp); err != nil {
		return err
	}

	return nil
}
