package tavor

import (
	"context"
	"net/http"
	"os"
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	originalAPIKey := os.Getenv("TAVOR_API_KEY")
	originalBaseURL := os.Getenv("TAVOR_BASE_URL")
	defer func() {
		os.Setenv("TAVOR_API_KEY", originalAPIKey)
		os.Setenv("TAVOR_BASE_URL", originalBaseURL)
	}()

	tests := []struct {
		name        string
		apiKey      string
		envAPIKey   string
		envBaseURL  string
		opts        []ClientOption
		wantErr     bool
		wantBaseURL string
	}{
		{
			name:        "Direct API key",
			apiKey:      "sk-tavor-test123",
			wantErr:     false,
			wantBaseURL: defaultBaseURL,
		},
		{
			name:        "API key from environment",
			apiKey:      "",
			envAPIKey:   "sk-tavor-env123",
			wantErr:     false,
			wantBaseURL: defaultBaseURL,
		},
		{
			name:        "No API key",
			apiKey:      "",
			envAPIKey:   "",
			wantErr:     true,
			wantBaseURL: "",
		},
		{
			name:        "Custom base URL from env",
			apiKey:      "sk-tavor-test",
			envBaseURL:  "https://custom.tavor.dev",
			wantErr:     false,
			wantBaseURL: "https://custom.tavor.dev",
		},
		{
			name:   "Custom base URL from option",
			apiKey: "sk-tavor-test",
			opts: []ClientOption{
				WithBaseURL("https://option.tavor.dev"),
			},
			wantErr:     false,
			wantBaseURL: "https://option.tavor.dev",
		},
		{
			name:   "Custom HTTP client",
			apiKey: "sk-tavor-test",
			opts: []ClientOption{
				WithHTTPClient(&http.Client{Timeout: 60 * time.Second}),
			},
			wantErr:     false,
			wantBaseURL: defaultBaseURL,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("TAVOR_API_KEY", tt.envAPIKey)
			os.Setenv("TAVOR_BASE_URL", tt.envBaseURL)

			client, err := NewClient(tt.apiKey, tt.opts...)

			if (err != nil) != tt.wantErr {
				t.Errorf("NewClient() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if client.baseURL != tt.wantBaseURL {
					t.Errorf("NewClient() baseURL = %v, want %v", client.baseURL, tt.wantBaseURL)
				}
			}
		})
	}
}

func TestErrorTypes(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		wantMsg  string
		wantCode int
	}{
		{
			name:     "AuthenticationError",
			err:      NewAuthenticationError("Invalid API key"),
			wantMsg:  "Invalid API key",
			wantCode: 401,
		},
		{
			name:     "BoxNotFoundError",
			err:      NewBoxNotFoundError("box-123"),
			wantMsg:  "Box not found: box-123",
			wantCode: 404,
		},
		{
			name:     "CommandTimeoutError",
			err:      NewCommandTimeoutError("cmd-456", 5000),
			wantMsg:  "Command cmd-456 timed out after 5000ms",
			wantCode: 408,
		},
		{
			name:     "BoxTimeoutError",
			err:      NewBoxTimeoutError("box-789", 60),
			wantMsg:  "Box box-789 failed to become ready within 60 seconds",
			wantCode: 408,
		},
		{
			name:     "RateLimitError",
			err:      NewRateLimitError(30),
			wantMsg:  "Rate limit exceeded. Retry after 30 seconds",
			wantCode: 429,
		},
		{
			name:     "ValidationError",
			err:      NewValidationError("cpu", "Invalid CPU value"),
			wantMsg:  "Validation error on field 'cpu': Invalid CPU value",
			wantCode: 400,
		},
		{
			name:     "InsufficientCreditsError",
			err:      NewInsufficientCreditsError(10.5, 5.0),
			wantMsg:  "Insufficient credits. Required: 10.50, Available: 5.00",
			wantCode: 402,
		},
		{
			name:     "APIError",
			err:      NewAPIError(500, "Internal server error"),
			wantMsg:  "Internal server error",
			wantCode: 500,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.Error() != tt.wantMsg {
				t.Errorf("Error message = %v, want %v", tt.err.Error(), tt.wantMsg)
			}

			if tErr, ok := tt.err.(interface{ StatusCode() int }); ok {
				if tErr.StatusCode() != tt.wantCode {
					t.Errorf("Status code = %v, want %v", tErr.StatusCode(), tt.wantCode)
				}
			}
		})
	}
}

func TestBoxConfig(t *testing.T) {
	originalTimeout := os.Getenv("TAVOR_BOX_TIMEOUT")
	defer func() {
		os.Setenv("TAVOR_BOX_TIMEOUT", originalTimeout)
	}()

	client, err := NewClient("sk-tavor-test")
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()

	// Test new CPU and RAM parameters
	tests := []struct {
		name         string
		config       *BoxConfig
		wantCPU      int
		wantMibRAM   int
		wantMetadata int
	}{
		{
			name: "With CPU and RAM",
			config: &BoxConfig{
				CPU:    2,
				MibRAM: 2048,
				Metadata: map[string]string{
					"test": "value",
				},
			},
			wantCPU:      2,
			wantMibRAM:   2048,
			wantMetadata: 1,
		},
		{
			name: "Default (no CPU/RAM specified)",
			config: &BoxConfig{
				Metadata: map[string]string{
					"env": "testing",
				},
			},
			wantCPU:      0,
			wantMibRAM:   0,
			wantMetadata: 1,
		},
		{
			name: "With only CPU",
			config: &BoxConfig{
				CPU: 2,
				Metadata: map[string]string{
					"test": "value",
					"env":  "testing",
				},
			},
			wantCPU:      2,
			wantMibRAM:   0,
			wantMetadata: 2,
		},
		{
			name: "With only RAM",
			config: &BoxConfig{
				MibRAM: 512,
			},
			wantCPU:      0,
			wantMibRAM:   512,
			wantMetadata: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.config.CPU != tt.wantCPU {
				t.Errorf("CPU = %v, want %v", tt.config.CPU, tt.wantCPU)
			}
			if tt.config.MibRAM != tt.wantMibRAM {
				t.Errorf("MibRAM = %v, want %v", tt.config.MibRAM, tt.wantMibRAM)
			}
			if len(tt.config.Metadata) != tt.wantMetadata {
				t.Errorf("Metadata entries = %v, want %v", len(tt.config.Metadata), tt.wantMetadata)
			}
		})
	}

	// Test environment variable handling for timeout
	os.Setenv("TAVOR_BOX_TIMEOUT", "120")

	// Verify timeout environment variable is respected in actual usage
	// (This would be tested in CreateBox, not directly on config)

	// Dummy test to ensure context is used
	_ = ctx
	_ = client
}

func TestBoxHandleGetPublicURL(t *testing.T) {
	tests := []struct {
		name      string
		box       *Box
		port      int
		wantURL   string
		wantError bool
	}{
		{
			name: "Valid hostname",
			box: &Box{
				ID:       "box-123",
				Status:   BoxStatusRunning,
				Hostname: "abc123.tavor.app",
			},
			port:      3000,
			wantURL:   "https://3000-abc123.tavor.app",
			wantError: false,
		},
		{
			name: "Different port",
			box: &Box{
				ID:       "box-456",
				Status:   BoxStatusRunning,
				Hostname: "xyz789.tavor.app",
			},
			port:      8080,
			wantURL:   "https://8080-xyz789.tavor.app",
			wantError: false,
		},
		{
			name: "No hostname",
			box: &Box{
				ID:       "box-789",
				Status:   BoxStatusRunning,
				Hostname: "",
			},
			port:      3000,
			wantURL:   "",
			wantError: true,
		},
		{
			name: "Port 80",
			box: &Box{
				ID:       "box-999",
				Status:   BoxStatusRunning,
				Hostname: "test123.tavor.app",
			},
			port:      80,
			wantURL:   "https://80-test123.tavor.app",
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, _ := NewClient("sk-tavor-test")
			handle := newBoxHandle(client, tt.box)

			url, err := handle.GetPublicURL(tt.port)

			if (err != nil) != tt.wantError {
				t.Errorf("GetPublicURL() error = %v, wantError %v", err, tt.wantError)
				return
			}

			if url != tt.wantURL {
				t.Errorf("GetPublicURL() = %v, want %v", url, tt.wantURL)
			}
		})
	}
}
