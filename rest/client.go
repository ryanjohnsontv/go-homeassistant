// A Go client for communicating with Home Assistant's REST API.
// https://developers.home-assistant.io/docs/api/rest

package rest

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/ryanjohnsontv/go-homeassistant/shared"
	"github.com/ryanjohnsontv/go-homeassistant/shared/constants/domains"
	"github.com/ryanjohnsontv/go-homeassistant/shared/types"
)

type (
	Client struct {
		apiURL      string // Formatted Home Assistant REST API URL (http://ha.local:8123/api)
		secure      bool   // When true uses https instead of http
		bearerToken string // Long-Lived Token from Home Assistant
		httpClient  *http.Client
	}

	ClientOption func(*Client)
)

func WithSecureConnection() ClientOption {
	return func(c *Client) {
		c.secure = true
	}
}

func NewClient(host, accessToken string, options ...ClientOption) (*Client, error) {
	if host == "" {
		return nil, ErrMissingHAAddress
	}

	if accessToken == "" {
		return nil, ErrMissingToken
	}

	apiURL := url.URL{Host: host, Path: "/api/", Scheme: "http"}

	c := &Client{
		bearerToken: "Bearer " + accessToken,
		httpClient:  http.DefaultClient,
	}

	for _, option := range options {
		option(c)
	}

	if c.secure {
		apiURL.Scheme = "https"
	}

	c.apiURL = apiURL.String()

	return c, nil
}

func (c *Client) newGETRequest(path string) (*http.Request, error) {
	return http.NewRequest(http.MethodGet, c.apiURL+path, nil)
}

func (c *Client) newPOSTRequest(path string, payload any) (*http.Request, error) {
	if payload == nil {
		return http.NewRequest(http.MethodPost, c.apiURL+path, http.NoBody)
	}

	b, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	return http.NewRequest(http.MethodPost, c.apiURL+path, bytes.NewBuffer(b))
}

// SendRequest sends an HTTP request and returns the response.
// Pass through pointer to decode a JSON response.
func (c *Client) sendRequest(req *http.Request, body any) error {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", c.bearerToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}

		return nil
	}

	if resp.StatusCode == http.StatusBadRequest ||
		resp.StatusCode == http.StatusUnauthorized ||
		resp.StatusCode == http.StatusNotFound ||
		resp.StatusCode == http.StatusMethodNotAllowed {
		var errResp apiResponse
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
			return fmt.Errorf("%d: unable to decode error response", resp.StatusCode)
		}

		return errors.New(errResp.Message)
	}

	if resp.StatusCode >= 400 {
		return fmt.Errorf("%d: unexpected error", resp.StatusCode)
	}

	return nil
}

type apiResponse struct {
	Message string `json:"message"`
}

// GetHealth returns an error if the API is unhealthy.
func (c *Client) GetHealth() error {
	req, err := c.newGETRequest("")
	if err != nil {
		return err
	}

	var resp apiResponse
	if err = c.sendRequest(req, &resp); err != nil {
		return err
	}

	return nil
}

// GetConfig gets the Home Assistant configuration.
func (c *Client) GetConfig() (types.HassConfig, error) {
	req, err := c.newGETRequest("config")
	if err != nil {
		return types.HassConfig{}, err
	}

	var resp types.HassConfig
	if err = c.sendRequest(req, &resp); err != nil {
		return types.HassConfig{}, err
	}

	return resp, nil
}

type Event struct {
	Event         string `json:"event"`
	ListenerCount int    `json:"listener_count"`
}

// GetEvents gets a list of all events in Home Assistant.
func (c *Client) GetEvents() ([]Event, error) {
	req, err := c.newGETRequest("events")
	if err != nil {
		return nil, err
	}

	var resp []Event
	if err = c.sendRequest(req, &resp); err != nil {
		return nil, err
	}

	return resp, nil
}

type Services struct {
	Domain   domains.Domain `json:"domain"`
	Services map[string]struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Fields      map[string]struct {
			Default     any            `json:"default,omitempty"`
			Description string         `json:"description,omitempty"`
			Example     any            `json:"example,omitempty"`
			Filter      map[string]any `json:"filter,omitempty"`
			Name        string         `json:"name,omitempty"`
			Required    bool           `json:"required,omitempty"`
			Selector    map[string]any `json:"selector,omitempty"`
		} `json:"fields"`
		Response map[string]any `json:"response,omitempty"`
		Target   map[string]any `json:"target,omitempty"`
	} `json:"services"`
}

// GetServices gets a list of all services in Home Assistant.
func (c *Client) GetServices() ([]Services, error) {
	req, err := c.newGETRequest("services")
	if err != nil {
		return nil, err
	}

	var resp []Services
	if err = c.sendRequest(req, &resp); err != nil {
		return nil, err
	}

	return resp, nil
}

type (
	GetHistoryOptions struct {
		Timestamp              time.Time
		EndTime                time.Time
		MinimalResponse        bool
		NoAttributes           bool
		SignificantChangesOnly bool
	}
	History []types.HassEntity
)

// GetHistory gets the history of events and state changes in Home Assistant.
func (c *Client) GetHistory(entityIDs []string, opts GetHistoryOptions) (History, error) {
	path := "history/period"
	if !opts.Timestamp.IsZero() {
		path += "/" + url.QueryEscape(opts.Timestamp.Format(time.RFC3339))
	}

	if len(entityIDs) > 0 {
		path += "?filter_entity_id=" + strings.Join(entityIDs, ",")
	}

	if !opts.EndTime.IsZero() {
		path += "?end_time=" + url.QueryEscape(opts.EndTime.Format(time.RFC3339))
	}

	if opts.MinimalResponse {
		path += "?minimal_response"
	}

	if opts.NoAttributes {
		path += "?no_attributes"
	}

	if opts.SignificantChangesOnly {
		path += "?significant_changes_only"
	}

	req, err := c.newGETRequest(path)
	if err != nil {
		return nil, err
	}

	var resp History
	if err = c.sendRequest(req, &resp); err != nil {
		return nil, err
	}

	return resp, nil
}

type (
	LogbookEntry struct {
		ContextUserID *string         `json:"context_user_id"`
		Domain        domains.Domain  `json:"domain"`
		EntityID      shared.EntityID `json:"entity_id"`
		Message       string          `json:"message"`
		Name          string          `json:"name"`
		When          time.Time       `json:"when"`
	}
	GetLogbookOptions struct {
		Timestamp time.Time
		EndTime   time.Time
		EntityID  string
	}
)

// GetLogbook gets the logbook of events in Home Assistant.
func (c *Client) GetLogbook(opts GetLogbookOptions) ([]LogbookEntry, error) {
	path := "logbook"
	if !opts.Timestamp.IsZero() {
		path += "/" + opts.Timestamp.Format(time.RFC3339)
	}

	req, err := c.newGETRequest(path)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	if !opts.EndTime.IsZero() {
		q.Add("end_time", opts.EndTime.Format(time.RFC3339))
	}

	if opts.EntityID != "" {
		q.Add("entity", opts.EntityID)
	}

	var resp []LogbookEntry
	if err = c.sendRequest(req, &resp); err != nil {
		return nil, err
	}

	return resp, nil
}

// GetStates gets a list of all states in Home Assistant.
func (c *Client) GetStates() ([]types.HassEntity, error) {
	req, err := c.newGETRequest("states")
	if err != nil {
		return nil, err
	}

	var resp []types.HassEntity
	if err = c.sendRequest(req, &resp); err != nil {
		return nil, err
	}

	return resp, nil
}

// GetState gets the state of an entity in Home Assistant.
func (c *Client) GetState(entityID string) (types.HassEntity, error) {
	req, err := c.newGETRequest("states/" + entityID)
	if err != nil {
		return types.HassEntity{}, err
	}

	var resp types.HassEntity
	if err = c.sendRequest(req, &resp); err != nil {
		return types.HassEntity{}, err
	}

	return resp, nil
}

// GetErrorLog gets the error log in Home Assistant.
func (c *Client) GetErrorLog() ([]map[string]any, error) {
	req, err := c.newGETRequest("error_log")
	if err != nil {
		return nil, err
	}

	var errorLog []map[string]any
	if err = c.sendRequest(req, &errorLog); err != nil {
		return nil, err
	}

	return errorLog, nil
}

// GetCameraProxy gets a proxy URL for a camera in Home Assistant.
func (c *Client) GetCameraProxy(entityID string) (string, error) {
	req, err := c.newGETRequest("camera_proxy/" + entityID)
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

type Calendars struct {
	EntityID shared.EntityID `json:"entity_id"`
	Name     string          `json:"name"`
}

// GetCalendars gets a list of calendar entities in Home Assistant.
func (c *Client) GetCalendars(calendarID string) ([]Calendars, error) {
	req, err := c.newGETRequest("calendars/" + calendarID)
	if err != nil {
		return nil, err
	}

	var resp []Calendars
	if err = c.sendRequest(req, &resp); err != nil {
		return nil, err
	}

	return resp, nil
}

type calendarDate struct {
	Date     string    `json:"date"`
	DateTime time.Time `json:"dateTime"`
}

type CalendarEvents []struct {
	Summary     string       `json:"summary"`
	Start       calendarDate `json:"start"`
	End         calendarDate `json:"end"`
	Description string       `json:"description"`
	Location    string       `json:"location"`
}

// GetCalendarEvents gets the events of a calendar in Home Assistant.
func (c *Client) GetCalendarEvents(calendarID string, start *time.Time, end *time.Time) (CalendarEvents, error) {
	path := "calendars/" + calendarID
	if start != nil {
		path += "?start=" + start.Format(time.RFC3339)
	}

	if end != nil {
		path += "?end=" + end.Format(time.RFC3339)
	}

	req, err := c.newGETRequest(path)
	if err != nil {
		return nil, err
	}

	var resp CalendarEvents
	if err = c.sendRequest(req, &resp); err != nil {
		return nil, err
	}

	return resp, nil
}

type UpsertStateRequest struct {
	EntityID   string `json:"-"`
	State      any    `json:"state"`
	Attributes any    `json:"attributes,omitempty"`
}

// UpsertState updates or creates a state in Home Assistant.
// Returns a state object and a URL of the new resource if one is created.
func (c *Client) UpsertState(params UpsertStateRequest) (types.HassEntity, *url.URL, error) {
	req, err := c.newPOSTRequest("states/"+params.EntityID, params)
	if err != nil {
		return types.HassEntity{}, nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return types.HassEntity{}, nil, err
	}

	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		var v types.HassEntity
		if err := json.NewDecoder(resp.Body).Decode(&v); err != nil {
			return types.HassEntity{}, nil, err
		}

		return v, nil, err
	case http.StatusCreated:
		var v types.HassEntity
		if err := json.NewDecoder(resp.Body).Decode(&v); err != nil {
			return types.HassEntity{}, nil, err
		}

		newResource, err := url.Parse(resp.Header.Get("Location"))

		return v, newResource, err
	default:
		return types.HassEntity{}, nil, fmt.Errorf("failed to upsert state")
	}
}

// FireEvent fires an event in Home Assistant.
// Returns a message if successful.
func (c *Client) FireEvent(eventType string, eventData any) (string, error) {
	req, err := c.newPOSTRequest("events/"+eventType, eventData)
	if err != nil {
		return "", err
	}

	var resp apiResponse
	if err = c.sendRequest(req, &resp); err != nil {
		return "", err
	}

	return resp.Message, nil
}

type CallServiceParams struct {
	Domain  domains.Domain
	Service string
}

// CallService calls a Home Assistant service via the REST API.
// Returns a list of states that have changed while the service was being executed.
func (c *Client) CallService(domain domains.Domain, service string, data any) ([]types.HassEntity, error) {
	req, err := c.newPOSTRequest("services/"+domain.String()+"/"+service, data)
	if err != nil {
		return nil, err
	}

	var resp []types.HassEntity
	if err = c.sendRequest(req, &resp); err != nil {
		return nil, err
	}

	return resp, nil
}

type Template struct {
	Template  string         `json:"template"`
	Variables map[string]any `json:"variable,omitempty"`
}

// RenderTemplate renders a Home Assistant template.
func (c *Client) RenderTemplate(template Template) (string, error) {
	req, err := c.newPOSTRequest("template", template)
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
	Result   string  `json:"result"`
	Errors   *string `json:"errors,omitempty"`
	Warnings *string `json:"warnings,omitempty"`
}

// CheckConfig checks the Home Assistant configuration.
// If the checks is successful, nil will be returned.
// If the check fails, a string containing the error will be returned.
func (c *Client) CheckConfig() error {
	req, err := c.newPOSTRequest("config/core/check_config", nil)
	if err != nil {
		return err
	}

	var resp checkConfig
	if err = c.sendRequest(req, &resp); err != nil {
		return err
	}

	if resp.Result != "valid" {
		return errors.New(*resp.Errors)
	}

	return nil
}

// HandleIntent handles an intent in Home Assistant.
// You must add intent: to your Home Assistant configuration file to enable this endpoint.
func (c *Client) HandleIntent(intent any) error {
	req, err := c.newPOSTRequest("services/intent/handle", intent)
	if err != nil {
		return err
	}

	if err = c.sendRequest(req, nil); err != nil {
		return err
	}

	return nil
}
