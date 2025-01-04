package websocket

import (
	"errors"
	"net/url"
	"regexp"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/ryanjohnsontv/go-homeassistant/logging"
	"github.com/ryanjohnsontv/go-homeassistant/shared/entity"
	"github.com/ryanjohnsontv/go-homeassistant/shared/types"
	"github.com/ryanjohnsontv/go-homeassistant/shared/version"
)

type Client struct {
	wsURL                   string // Formatted Home Assistant websocket URL (ws://ha.local:8123/api/websocket)
	accessToken             string // Long-Lived Token from Home Assistant
	secure                  bool
	haVersion               version.Version
	wsConn                  *websocket.Conn
	timeout                 time.Duration
	logger                  logging.Logger
	msgID                   int64
	eventHandler            map[int64]eventHandler
	triggerHandler          map[int64][]triggerHandler
	entityListeners         map[entity.ID][]entityListener
	regexEntityListeners    map[*regexp.Regexp][]entityListener
	dateTimeEntityListeners map[time.Time]map[entity.ID][]dateTimeEntityTrigger
	resultChan              map[int64]chan []byte
	pongChan                chan bool
	stopChan                chan bool
	reconnectChan           chan bool
	mu                      sync.Mutex
	msgHistory              map[int64]cmdMessage
	EntitiesMap             types.EntitiesMap
}

type (
	ClientOption func(*Client)

	eventHandler struct {
		EventType string
		Callback  func(types.Event)
	}
	triggerHandler struct {
		Callback func(types.Trigger)
	}
	entityListener struct {
		callback      func(*types.StateChange)
		FilterOptions filterOptions
	}
)

func NewClient(host, accessToken string, options ...ClientOption) (*Client, error) {
	if host == "" {
		return nil, errors.New("home assistant address is required")
	}

	if accessToken == "" {
		return nil, errors.New("access token is required")
	}

	wsURL := url.URL{Host: host, Path: "/api/websocket", Scheme: "ws"}

	c := &Client{
		accessToken:             accessToken,
		timeout:                 10 * time.Second,
		logger:                  &logging.DefaultLogger{},
		eventHandler:            make(map[int64]eventHandler),
		triggerHandler:          make(map[int64][]triggerHandler),
		entityListeners:         make(map[entity.ID][]entityListener),
		regexEntityListeners:    make(map[*regexp.Regexp][]entityListener),
		dateTimeEntityListeners: make(map[time.Time]map[entity.ID][]dateTimeEntityTrigger),
		resultChan:              make(map[int64]chan []byte),
		pongChan:                make(chan bool),
		stopChan:                make(chan bool),
		reconnectChan:           make(chan bool),
		EntitiesMap:             make(types.EntitiesMap),
		msgHistory:              make(map[int64]cmdMessage),
	}

	for _, option := range options {
		option(c)
	}

	if c.secure {
		wsURL.Scheme = "wss"
	}

	c.wsURL = wsURL.String()

	return c, c.run()
}

func WithCustomLogger(logger logging.Logger) ClientOption {
	return func(c *Client) {
		c.logger = logger
	}
}

func WithSecureConnection() ClientOption {
	return func(c *Client) {
		c.secure = true
	}
}

func (c *Client) run() error {
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
