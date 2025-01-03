package websocket

import (
	"net/url"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/ryanjohnsontv/go-homeassistant/logging"
	"github.com/ryanjohnsontv/go-homeassistant/shared/types"
)

type Client struct {
	wsURL                   url.URL // Formatted Home Assistant websocket URL (ws://ha.local:8123/api/websocket)
	accessToken             string  // Long-Lived Token from Home Assistant
	haVersion               string
	wsConn                  *websocket.Conn
	logger                  logging.Logger
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
	msgHistory              map[int64]cmdMessage
	stateVars               map[string]any
	States                  map[string]State
}

type Config struct {
	Host        string // The host:port of your home assistant instance. (ex: homeassistant.local:8123)
	AccessToken string // The log-lived access token generated in Home Assistant
}

type (
	ClientOption func(*Client)

	eventHandler struct {
		EventType string
		Callback  func(*types.HassEvent)
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

func NewClient(cfg Config, options ...ClientOption) (*Client, error) {
	if cfg.Host == "" {
		return nil, ErrMissingHAAddress
	}

	if cfg.AccessToken == "" {
		return nil, ErrMissingToken
	}

	c := &Client{
		wsURL:                   url.URL{Host: cfg.Host, Path: "/api/websocket", Scheme: "ws"},
		accessToken:             cfg.AccessToken,
		logger:                  &logging.DefaultLogger{},
		eventHandler:            make(map[int64]eventHandler),
		triggerHandler:          make(map[int64][]triggerHandler),
		entityListeners:         make(map[string][]entityListener),
		regexEntityListeners:    make(map[string][]entityListener),
		dateTimeEntityListeners: make(map[time.Time]map[string][]dateTimeEntityTrigger),
		resultChan:              make(map[int64]chan []byte),
		pongChan:                make(chan bool),
		stopChan:                make(chan bool),
		reconnectChan:           make(chan bool),
		States:                  make(map[string]State),
		msgHistory:              make(map[int64]cmdMessage),
	}

	for _, option := range options {
		option(c)
	}

	return c, nil
}

func WithCustomLogger(logger logging.Logger) ClientOption {
	return func(c *Client) {
		c.logger = logger
	}
}

func WithSecureConnection() ClientOption {
	return func(c *Client) {
		c.wsURL.Scheme = "wss"
	}
}

func WithCustomStateVars(states map[string]any) ClientOption {
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
