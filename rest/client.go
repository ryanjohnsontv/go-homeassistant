// A Go client for communicating with Home Assistant's REST API.
// https://developers.home-assistant.io/docs/api/rest
package rest

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/ryanjohnsontv/go-homeassistant/logging"
	"github.com/ryanjohnsontv/go-homeassistant/shared/constants/domains"
	"github.com/ryanjohnsontv/go-homeassistant/shared/types"
	"github.com/ryanjohnsontv/go-homeassistant/shared/utils"
)

type (
	Client struct {
		apiURL           *url.URL // Formatted Home Assistant REST API URL (http://ha.local:8123/api)
		apiURLString     string
		bearerToken      string          // Long-Lived Token from Home Assistant
		httpClient       *http.Client    // HTTP client
		streamHTTPClient *http.Client    // Client for event streams
		ctx              context.Context // Defaults to context.Background()
		logger           logging.Logger  // Optional custom logger
	}

	ClientOption func(*Client)
)

func NewClient(host, accessToken string, options ...ClientOption) (*Client, error) {
	if host == "" {
		return nil, errors.New("home assistant address is required")
	}

	if accessToken == "" {
		return nil, errors.New("access token is required")
	}

	apiURL, err := utils.GetHTTPURL(host)
	if err != nil {
		return nil, fmt.Errorf("invalid home assistant host: %w", err)
	}

	c := &Client{
		apiURL:      apiURL,
		bearerToken: "Bearer " + accessToken,
		ctx:         context.Background(),
		logger:      logging.NewLogger(),
		httpClient: &http.Client{
			Transport: &http.Transport{
				MaxIdleConns:       100,
				IdleConnTimeout:    90 * time.Second,
				DisableCompression: false,
			},
			Timeout: 10 * time.Second,
		},
		streamHTTPClient: &http.Client{
			Transport: &http.Transport{
				MaxIdleConns:       100,
				IdleConnTimeout:    90 * time.Second,
				DisableCompression: false,
			},
			Timeout: 0,
		},
	}

	for _, option := range options {
		option(c)
	}

	c.apiURLString = c.apiURL.String()

	c.logger.Debug("using %s as api url", c.apiURL)

	return c, nil
}

func WithAPIPath(path string) ClientOption {
	return func(c *Client) {
		c.apiURL.Path = path
	}
}

func WithContext(ctx context.Context) ClientOption {
	return func(c *Client) {
		c.ctx = ctx
	}
}

func WithHTTPClient(client *http.Client) ClientOption {
	return func(c *Client) {
		c.httpClient = client
	}
}

func WithLogger(logger logging.Logger) ClientOption {
	return func(c *Client) {
		c.logger = logger
	}
}

func WithSecureConnection() ClientOption {
	return func(c *Client) {
		c.apiURL.Scheme = "https"
	}
}

func WithStreamHTTPClient(client *http.Client) ClientOption {
	return func(c *Client) {
		c.streamHTTPClient = client
	}
}

func WithTimeout(timeout time.Duration) ClientOption {
	return func(c *Client) {
		if c.httpClient == nil {
			c.httpClient = http.DefaultClient
		}

		c.httpClient.Timeout = timeout
	}
}

func (c *Client) newRequest(ctx context.Context, method, path string, body any) (*http.Request, error) {
	fullURL := c.apiURL.ResolveReference(&url.URL{Path: path}).String()

	var bodyReader io.Reader

	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}

		bodyReader = bytes.NewBuffer(data)

		c.logger.Debug("sending %s request to %s with body: %+v", method, fullURL, body)
	} else {
		bodyReader = http.NoBody

		c.logger.Debug("sending %s request to %s with no body", method, fullURL)
	}

	req, err := http.NewRequestWithContext(ctx, method, fullURL, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", c.bearerToken)

	return req, nil
}

// SendRequest sends an HTTP request and returns the response.
// Pass through pointer to decode a JSON response.
func (c *Client) sendRequest(req *http.Request, body any) error {
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
		return fmt.Errorf("%d: %s", resp.StatusCode, resp.Status)
	}

	return nil
}

type apiResponse struct {
	Message string `json:"message"`
}

// GetHealth returns an error if the API is unhealthy.
func (c *Client) GetHealth() error {
	return c.GetHealthWithContext(c.ctx)
}

// GetHealthWithContext returns an error if the API is unhealthy.
func (c *Client) GetHealthWithContext(ctx context.Context) error {
	req, err := c.newRequest(ctx, http.MethodGet, "", nil)
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
func (c *Client) GetConfig() (types.Config, error) {
	return c.GetConfigWithContext(c.ctx)
}

// GetConfigWithContext gets the Home Assistant configuration.
func (c *Client) GetConfigWithContext(ctx context.Context) (types.Config, error) {
	req, err := c.newRequest(ctx, http.MethodGet, "config", nil)
	if err != nil {
		return types.Config{}, err
	}

	var resp types.Config
	if err = c.sendRequest(req, &resp); err != nil {
		return types.Config{}, err
	}

	return resp, nil
}

type Event struct {
	Event         string `json:"event"`
	ListenerCount int    `json:"listener_count"`
}

// GetEvents gets a list of all events in Home Assistant.
func (c *Client) GetEvents() ([]Event, error) {
	return c.GetEventsWithContext(c.ctx)
}

// GetEventsWithContext gets a list of all events in Home Assistant.
func (c *Client) GetEventsWithContext(ctx context.Context) ([]Event, error) {
	req, err := c.newRequest(ctx, http.MethodGet, "events", nil)
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
			Default     any            `json:"default"`
			Description string         `json:"description"`
			Example     any            `json:"example"`
			Filter      map[string]any `json:"filter"`
			Name        string         `json:"name"`
			Required    bool           `json:"required"`
			Selector    map[string]any `json:"selector"`
		} `json:"fields"`
		Response map[string]any `json:"response"`
		Target   map[string]any `json:"target"`
	} `json:"services"`
}

// GetServices gets a list of all services in Home Assistant.
func (c *Client) GetServices() ([]Services, error) {
	return c.GetServicesWithContext(c.ctx)
}

// GetServicesWithContext gets a list of all services in Home Assistant.
func (c *Client) GetServicesWithContext(ctx context.Context) ([]Services, error) {
	req, err := c.newRequest(ctx, http.MethodGet, "services", nil)
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
	History []types.Entity
)

// GetHistory gets the history of events and state changes in Home Assistant.
func (c *Client) GetHistory(
	entityIDs []string,
	opts GetHistoryOptions,
) (History, error) {
	return c.GetHistoryWithContext(c.ctx, entityIDs, opts)
}

// GetHistoryWithContext gets the history of events and state changes in Home Assistant.
func (c *Client) GetHistoryWithContext(
	ctx context.Context,
	entityIDs []string,
	opts GetHistoryOptions,
) (History, error) {
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

	req, err := c.newRequest(ctx, http.MethodGet, path, nil)
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
		ContextUserID *string        `json:"context_user_id"`
		Domain        domains.Domain `json:"domain"`
		EntityID      string         `json:"entity_id"`
		Message       string         `json:"message"`
		Name          string         `json:"name"`
		When          time.Time      `json:"when"`
	}
	GetLogbookOptions struct {
		Timestamp time.Time
		EndTime   time.Time
		EntityID  string
	}
)

// GetLogbook gets the logbook of events in Home Assistant.
func (c *Client) GetLogbook(opts GetLogbookOptions) ([]LogbookEntry, error) {
	return c.GetLogbookWithContext(c.ctx, opts)
}

// GetLogbookWithContext gets the logbook of events in Home Assistant.
func (c *Client) GetLogbookWithContext(
	ctx context.Context,
	opts GetLogbookOptions,
) ([]LogbookEntry, error) {
	path := "logbook"
	if !opts.Timestamp.IsZero() {
		path += "/" + opts.Timestamp.Format(time.RFC3339)
	}

	req, err := c.newRequest(ctx, http.MethodGet, path, nil)
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
func (c *Client) GetStates() ([]types.Entity, error) {
	return c.GetStatesWithContext(c.ctx)
}

// GetStatesWithContext gets a list of all states in Home Assistant.
func (c *Client) GetStatesWithContext(ctx context.Context) ([]types.Entity, error) {
	req, err := c.newRequest(ctx, http.MethodGet, "states", nil)
	if err != nil {
		return nil, err
	}

	var resp []types.Entity
	if err = c.sendRequest(req, &resp); err != nil {
		return nil, err
	}

	return resp, nil
}

// GetState gets the state of an entity in Home Assistant.
func (c *Client) GetState(entityID string) (types.Entity, error) {
	return c.GetStateWithContext(c.ctx, entityID)
}

// GetStateWithContext gets the state of an entity in Home Assistant.
func (c *Client) GetStateWithContext(ctx context.Context, entityID string) (types.Entity, error) {
	req, err := c.newRequest(ctx, http.MethodGet, "states/"+entityID, nil)
	if err != nil {
		return types.Entity{}, err
	}

	var resp types.Entity
	if err = c.sendRequest(req, &resp); err != nil {
		return types.Entity{}, err
	}

	return resp, nil
}

// GetErrorLog gets the error log in Home Assistant.
func (c *Client) GetErrorLog() ([]map[string]any, error) {
	return c.GetErrorLogWithContext(c.ctx)
}

// GetErrorLogWithContext gets the error log in Home Assistant.
func (c *Client) GetErrorLogWithContext(ctx context.Context) ([]map[string]any, error) {
	req, err := c.newRequest(ctx, http.MethodGet, "error_log", nil)
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
	return c.GetCameraProxyWithContext(c.ctx, entityID)
}

// GetCameraProxyWithContext gets a proxy URL for a camera in Home Assistant.
func (c *Client) GetCameraProxyWithContext(ctx context.Context, entityID string) (string, error) {
	req, err := c.newRequest(ctx, http.MethodGet, "camera_proxy/"+entityID, nil)
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
	EntityID string `json:"entity_id"`
	Name     string `json:"name"`
}

// GetCalendars gets a list of calendar entities in Home Assistant.
func (c *Client) GetCalendars(calendarID string) ([]Calendars, error) {
	return c.GetCalendarsWithContext(c.ctx, calendarID)
}

// GetCalendarsWithContext gets a list of calendar entities in Home Assistant.
func (c *Client) GetCalendarsWithContext(ctx context.Context, calendarID string) ([]Calendars, error) {
	req, err := c.newRequest(ctx, http.MethodGet, "calendars/"+calendarID, nil)
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
func (c *Client) GetCalendarEvents(
	calendarID string,
	start *time.Time,
	end *time.Time,
) (CalendarEvents, error) {
	return c.GetCalendarEventsWithContext(c.ctx, calendarID, start, end)
}

// GetCalendarEventsWithContext gets the events of a calendar in Home Assistant.
func (c *Client) GetCalendarEventsWithContext(
	ctx context.Context,
	calendarID string,
	start *time.Time,
	end *time.Time,
) (CalendarEvents, error) {
	path := "calendars/" + calendarID
	if start != nil {
		path += "?start=" + start.Format(time.RFC3339)
	}

	if end != nil {
		path += "?end=" + end.Format(time.RFC3339)
	}

	req, err := c.newRequest(ctx, http.MethodGet, path, nil)
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
func (c *Client) UpsertState(params UpsertStateRequest) (types.Entity, *url.URL, error) {
	return c.UpsertStateWithContext(c.ctx, params)
}

// UpsertStateWithContext updates or creates a state in Home Assistant.
// Returns a state object and a URL of the new resource if one is created.
func (c *Client) UpsertStateWithContext(
	ctx context.Context,
	params UpsertStateRequest,
) (types.Entity, *url.URL, error) {
	req, err := c.newRequest(ctx, http.MethodPost, "states/"+params.EntityID, params)
	if err != nil {
		return types.Entity{}, nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return types.Entity{}, nil, err
	}

	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		var v types.Entity
		if err := json.NewDecoder(resp.Body).Decode(&v); err != nil {
			return types.Entity{}, nil, err
		}

		return v, nil, err
	case http.StatusCreated:
		var v types.Entity
		if err := json.NewDecoder(resp.Body).Decode(&v); err != nil {
			return types.Entity{}, nil, err
		}

		newResource, err := url.Parse(resp.Header.Get("Location"))

		return v, newResource, err
	default:
		return types.Entity{}, nil, fmt.Errorf("failed to upsert state")
	}
}

// FireEvent fires an event in Home Assistant.
// Returns a message if successful.
func (c *Client) FireEvent(eventType string, eventData any) (string, error) {
	return c.FireEventWithContext(c.ctx, eventType, eventData)
}

// FireEventWithContext fires an event in Home Assistant.
// Returns a message if successful.
func (c *Client) FireEventWithContext(ctx context.Context, eventType string, eventData any) (string, error) {
	req, err := c.newRequest(ctx, http.MethodPost, "events/"+eventType, eventData)
	if err != nil {
		return "", err
	}

	var resp apiResponse
	if err = c.sendRequest(req, &resp); err != nil {
		return "", err
	}

	return resp.Message, nil
}

type CallServiceData struct {
	EntityID string `json:"entity_id"`
}

// CallService calls a Home Assistant service via the REST API.
// Returns a list of states that have changed while the service was being executed.
func (c *Client) CallService(
	domain domains.Domain,
	service string,
	data any,
) ([]types.Entity, error) {
	return c.CallServiceWithContext(c.ctx, domain, service, data)
}

// CallServiceWithContext calls a Home Assistant service via the REST API.
// Returns a list of states that have changed while the service was being executed.
func (c *Client) CallServiceWithContext(
	ctx context.Context,
	domain domains.Domain,
	service string,
	data any,
) ([]types.Entity, error) {
	req, err := c.newRequest(ctx, http.MethodPost, "services/"+domain.String()+"/"+service, data)
	if err != nil {
		return nil, err
	}

	var resp []types.Entity
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
	return c.RenderTemplateWithContext(c.ctx, template)
}

// RenderTemplateWithContext renders a Home Assistant template.
func (c *Client) RenderTemplateWithContext(ctx context.Context, template Template) (string, error) {
	req, err := c.newRequest(ctx, http.MethodPost, "template", template)
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
	Errors   *string `json:"errors"`
	Warnings *string `json:"warnings"`
}

// CheckConfig checks the Home Assistant configuration.
// If the checks is successful, nil will be returned.
// If the check fails, a string containing the error will be returned.
func (c *Client) CheckConfig() error {
	return c.CheckConfigWithContext(c.ctx)
}

// CheckConfigWithContext checks the Home Assistant configuration.
// If the checks is successful, nil will be returned.
// If the check fails, a string containing the error will be returned.
func (c *Client) CheckConfigWithContext(ctx context.Context) error {
	req, err := c.newRequest(ctx, http.MethodPost, "config/core/check_config", nil)
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
	return c.HandleIntentWithContext(c.ctx, intent)
}

// HandleIntentWithContext handles an intent in Home Assistant.
// You must add intent: to your Home Assistant configuration file to enable this endpoint.
func (c *Client) HandleIntentWithContext(ctx context.Context, intent any) error {
	req, err := c.newRequest(ctx, http.MethodPost, "services/intent/handle", intent)
	if err != nil {
		return err
	}

	if err = c.sendRequest(req, nil); err != nil {
		return err
	}

	return nil
}

// EventStream connects to Home Assistant's event stream API and streams events.
// It writes events to the `events` channel and listens for a stop signal on `stop`.
// Each event is sent as a string in JSON format.
func (c *Client) EventStream(
	events chan<- string,
	stop <-chan struct{},
	restrictions ...string,
) error {
	return c.EventStreamWithContext(c.ctx, events, stop, restrictions...)
}

// EventStreamWithContext connects to Home Assistant's event stream API and streams events.
// It writes events to the `events` channel and listens for a stop signal on `stop`.
// Each event is sent as a string in JSON format.
func (c *Client) EventStreamWithContext(
	ctx context.Context,
	events chan<- string,
	stop <-chan struct{},
	restrictions ...string,
) error {
	path := "stream"
	if len(restrictions) > 0 {
		path += "?restrict=" + strings.Join(restrictions, ",")
	}

	req, err := c.newRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to connect to event stream: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected response status: %s", resp.Status)
	}

	reader := bufio.NewReader(resp.Body)

	for {
		select {
		case <-stop:
			return nil
		default:
			line, err := reader.ReadString('\n')
			if err != nil {
				return fmt.Errorf("error reading from event stream: %w", err)
			}

			events <- line
		}
	}
}
