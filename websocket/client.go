// A Go client for communicating with Home Assistant's Websocket API.
// https://developers.home-assistant.io/docs/api/websocket
package websocket

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"reflect"
	"regexp"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/ryanjohnsontv/go-homeassistant/logging"
	"github.com/ryanjohnsontv/go-homeassistant/shared/constants/domains"
	"github.com/ryanjohnsontv/go-homeassistant/shared/types"
	"github.com/ryanjohnsontv/go-homeassistant/shared/utils"
	"github.com/ryanjohnsontv/go-homeassistant/shared/version"
	"github.com/ryanjohnsontv/go-homeassistant/websocket/actions"
)

type Client struct {
	accessToken              string // Long-Lived Token from Home Assistant
	Actions                  *actions.Actions
	customEntityListeners    map[string][]customEntityListener
	customEntityPointers     map[string][]any        // Maps entity IDs to user-provided pointers
	customEntityTypes        map[string]reflect.Type // Maps entity IDs to custom types
	customEventHandler       map[int64]customEventHandler
	dialer                   *websocket.Dialer
	domainEntityListeners    map[domains.Domain][]entityListener
	EntitiesMap              types.EntitiesMap
	entityListeners          map[string][]entityListener
	eventHandler             map[int64]func(types.Event)
	haVersion                version.Version // Version of Home Assistant sent during auth phase
	logger                   logging.Logger
	msgHistory               map[int64]cmdMessage
	msgID                    int64
	mu                       sync.Mutex
	pongChan                 chan bool
	reconnectChan            chan bool
	regexEntityListeners     map[*regexp.Regexp][]entityListener
	resultChan               map[int64]chan []byte
	stopChan                 chan bool
	substringEntityListeners map[string][]entityListener
	timeout                  time.Duration //
	triggerHandler           map[int64][]triggerHandler
	wsConn                   *websocket.Conn // Websocket connection
	wsURL                    *url.URL        // Formatted Home Assistant websocket URL (ws://ha.local:8123/api/websocket)
}

type (
	ClientOption func(*Client)

	triggerHandler struct {
		Callback func(types.Trigger)
	}
	entityListener struct {
		callback      func(*types.StateChange)
		FilterOptions filterOptions
	}
	customEntityListener struct {
		callback func(oldState any, newState any)
	}
	customEventHandler struct {
		callback  func(any)
		eventType reflect.Type
	}
)

func NewClient(host, accessToken string, options ...ClientOption) (*Client, error) {
	if host == "" {
		return nil, errors.New("home assistant address is required")
	}

	if accessToken == "" {
		return nil, errors.New("access token is required")
	}

	wsURL, err := utils.GetWebsocketURL(host)
	if err != nil {
		return nil, fmt.Errorf("invalid home assistant host: %w", err)
	}

	c := &Client{
		accessToken:              accessToken,
		customEntityListeners:    make(map[string][]customEntityListener),
		customEntityPointers:     make(map[string][]any),
		customEntityTypes:        make(map[string]reflect.Type),
		customEventHandler:       make(map[int64]customEventHandler),
		dialer:                   websocket.DefaultDialer,
		domainEntityListeners:    make(map[domains.Domain][]entityListener),
		EntitiesMap:              make(types.EntitiesMap),
		entityListeners:          make(map[string][]entityListener),
		eventHandler:             make(map[int64]func(types.Event)),
		logger:                   logging.NewLogger(),
		msgHistory:               make(map[int64]cmdMessage),
		pongChan:                 make(chan bool),
		reconnectChan:            make(chan bool),
		regexEntityListeners:     make(map[*regexp.Regexp][]entityListener),
		resultChan:               make(map[int64]chan []byte),
		stopChan:                 make(chan bool),
		substringEntityListeners: make(map[string][]entityListener),
		timeout:                  10 * time.Second,
		triggerHandler:           make(map[int64][]triggerHandler),
		wsURL:                    wsURL,
	}

	for _, option := range options {
		option(c)
	}

	c.logger.Debug("using %s as websocket url", c.wsURL.String())

	c.Actions = actions.NewActionService(c)

	return c, nil
}

func WithCustomDialer(dialer *websocket.Dialer) ClientOption {
	return func(c *Client) {
		c.dialer = dialer
	}
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

	c.populateCustomPointers()

	return nil
}

func (c *Client) Close() {
	c.wsConn.Close()
}

func (c *Client) populateCustomPointers() {
	var wg sync.WaitGroup

	for entityID, val := range c.customEntityPointers {
		entity, exists := c.EntitiesMap[entityID]

		for _, v := range val {
			wg.Add(1)

			go func(entity types.Entity, pointer any) {
				defer wg.Done()

				if !exists {
					c.logger.Error(nil, "custom entity %s does not exist; cannot populate provided pointer")
					return
				}

				b, err := json.Marshal(entity)
				if err != nil {
					c.logger.Error(err, "unable to populate provided %s variable", entity.EntityID)
					return
				}

				if err := json.Unmarshal(b, pointer); err != nil {
					c.logger.Error(err, "unable to populate provided %s variable", entity.EntityID)
					return
				}
			}(entity, v)
		}
	}

	wg.Wait()
}
