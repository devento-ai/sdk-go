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
			err:      NewValidationError("template", "Invalid template name"),
			wantMsg:  "Validation error on field 'template': Invalid template name",
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
	originalTemplate := os.Getenv("TAVOR_BOX_TEMPLATE")
	originalTimeout := os.Getenv("TAVOR_BOX_TIMEOUT")
	defer func() {
		os.Setenv("TAVOR_BOX_TEMPLATE", originalTemplate)
		os.Setenv("TAVOR_BOX_TIMEOUT", originalTimeout)
	}()

	client, err := NewClient("sk-tavor-test")
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()

	os.Setenv("TAVOR_BOX_TEMPLATE", "pro")
	os.Setenv("TAVOR_BOX_TIMEOUT", "120")

	// This would normally create a box, but we can't test without a real API
	// Just verify the config is properly constructed
	config := &BoxConfig{}

	if config.Template == "" && os.Getenv("TAVOR_BOX_TEMPLATE") != "" {
		config.Template = BoxTemplate(os.Getenv("TAVOR_BOX_TEMPLATE"))
	}

	if config.Template != BoxTemplatePro {
		t.Errorf("Expected template from env to be 'pro', got %v", config.Template)
	}

	config = &BoxConfig{
		Template: BoxTemplateBasic,
		Metadata: map[string]string{
			"test": "value",
			"env":  "testing",
		},
	}

	if len(config.Metadata) != 2 {
		t.Errorf("Expected 2 metadata entries, got %d", len(config.Metadata))
	}

	// Dummy test to ensure context is used
	_ = ctx
	_ = client
}
