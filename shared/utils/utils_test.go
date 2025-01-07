package utils_test

import (
	"net/url"
	"testing"

	"github.com/ryanjohnsontv/go-homeassistant/shared/utils"
)

func TestGetURL(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		schema  string
		want    string
		wantErr bool
	}{
		{
			name:    "WebSocket: valid URL with port",
			input:   "ws://example.com:9000",
			schema:  "websocket",
			want:    "ws://example.com:9000/api/websocket",
			wantErr: false,
		},
		{
			name:    "WebSocket: add scheme and port",
			input:   "example.com",
			schema:  "websocket",
			want:    "ws://example.com:8123/api/websocket",
			wantErr: false,
		},
		{
			name:    "WebSocket: wrong scheme gets corrected",
			input:   "http://example.com",
			schema:  "websocket",
			want:    "ws://example.com:8123/api/websocket",
			wantErr: false,
		},
		{
			name:    "HTTP: valid URL",
			input:   "http://example.com",
			schema:  "http",
			want:    "http://example.com:8123/api/",
			wantErr: false,
		},
		{
			name:    "HTTP: missing scheme",
			input:   "example.com",
			schema:  "http",
			want:    "http://example.com:8123/api/",
			wantErr: false,
		},
		{
			name:    "Invalid URL",
			input:   "http://example:com",
			schema:  "http",
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got *url.URL
			var err error

			if tt.schema == "websocket" {
				got, err = utils.GetWebsocketURL(tt.input)
			} else {
				got, err = utils.GetHTTPURL(tt.input)
			}

			if (err != nil) != tt.wantErr {
				t.Errorf("unexpected error status: got %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err == nil && got.String() != tt.want {
				t.Errorf("unexpected result: got %s, want %s", got.String(), tt.want)
			}
		})
	}
}
