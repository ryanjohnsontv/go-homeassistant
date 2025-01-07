package utils

import (
	"fmt"
	"net/url"
	"strings"
)

type urlSchema string

const (
	websocket urlSchema = "ws"
	http      urlSchema = "http"
)

func GetWebsocketURL(rawURL string) (*url.URL, error) {
	return normalizeURL(rawURL, websocket)
}

func GetHTTPURL(rawURL string) (*url.URL, error) {
	return normalizeURL(rawURL, http)
}

func normalizeURL(rawURL string, scheme urlSchema) (*url.URL, error) {
	if !strings.Contains(rawURL, "://") {
		rawURL = string(scheme) + "://" + rawURL
	}

	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}

	parsedURL.Scheme = string(scheme)

	if parsedURL.Port() == "" {
		parsedURL.Host = fmt.Sprintf("%s:8123", parsedURL.Hostname())
	}

	switch scheme {
	case websocket:
		parsedURL.Path = "/api/websocket"
	case http:
		parsedURL.Path = "/api/"
	}

	return parsedURL, nil
}
