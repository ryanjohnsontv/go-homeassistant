package homeassistant

import (
	"net/http"
	"net/url"
	"os"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/rs/zerolog"
)

type Client struct {
	wsURL                   string
	apiURL                  string
	secure                  bool
	accessToken             string
	haVersion               string
	wsConn                  *websocket.Conn
	httpClient              *http.Client
	logger                  zerolog.Logger
	msgID                   int64
	eventHandler            map[int64]eventHandler
	triggerHandler          map[int64][]triggerHandler
	entityListeners         map[string][]entityListener
	regexEntityListeners    map[string][]entityListener
	dateTimeEntityListeners map[time.Time]map[string][]dateTimeEntityTrigger
	resultChan              map[int64]chan []byte
	pongChan                chan bool
	stopChan                chan bool
	reconnectChan           chan bool
	mu                      sync.Mutex
	msgHistory              map[int64]interface{}
	stateVars               map[string]interface{}
	States                  map[string]stateObj
}

type Config struct {
	Host        string // The host:port of your home assistant instance. (ex: homeassistant.local:8123)
	AccessToken string // The log-lived access token generated in Home Assistant
}

type (
	ClientOption func(*Client)

	eventHandler struct {
		EventType string
		Callback  func(*Event)
	}
	triggerHandler struct {
		Callback func(*Trigger)
	}
	entityListener struct {
		callback      func(*StateChange)
		FilterOptions filterOptions
	}
	dateTimeEntityTrigger struct {
		callback func()
		timeType string
	}
)

func NewWebsocketClient(cfg Config, options ...ClientOption) (*Client, error) {
	if cfg.Host == "" {
		return nil, ErrMissingHAAddress
	}
	if cfg.AccessToken == "" {
		return nil, ErrMissingToken
	}

	wsURL := url.URL{Host: cfg.Host, Path: "/api/websocket", Scheme: "ws"}
	apiURL := url.URL{Host: cfg.Host, Path: "/api/", Scheme: "http"}

	defaultLogger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr}).
		With().
		Timestamp().
		Logger().
		Level(zerolog.ErrorLevel)

	c := &Client{
		accessToken:             cfg.AccessToken,
		logger:                  defaultLogger,
		eventHandler:            make(map[int64]eventHandler),
		triggerHandler:          make(map[int64][]triggerHandler),
		entityListeners:         make(map[string][]entityListener),
		regexEntityListeners:    make(map[string][]entityListener),
		dateTimeEntityListeners: make(map[time.Time]map[string][]dateTimeEntityTrigger),
		resultChan:              make(map[int64]chan []byte),
		pongChan:                make(chan bool),
		stopChan:                make(chan bool),
		reconnectChan:           make(chan bool),
		States:                  make(map[string]stateObj),
		msgHistory:              make(map[int64]interface{}),
		httpClient:              http.DefaultClient,
	}

	for _, option := range options {
		option(c)
	}

	if c.secure {
		wsURL.Scheme = "wss"
		apiURL.Scheme = "https"
	}

	c.apiURL = apiURL.String()
	c.wsURL = wsURL.String()

	return c, nil
}

func WithCustomLogger(logger zerolog.Logger) ClientOption {
	return func(c *Client) {
		c.logger = logger
	}
}

func WithSecureConnection() ClientOption {
	return func(c *Client) {
		c.secure = true
	}
}

func WithCustomStateVars(states map[string]interface{}) ClientOption {
	return func(c *Client) {
		c.stateVars = states
	}
}

func (c *Client) Run() error {
	if err := c.connect(); err != nil {
		return err
	}
	go c.listen()
	go c.startHeartbeat()
	_, err := c.GetStates()
	if err != nil {
		return err
	}
	err = c.SubscribeToEvent("state_changed", nil)
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) Close() {
	c.wsConn.Close()
}
