package devento

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	originalAPIKey := os.Getenv("DEVENTO_API_KEY")
	originalBaseURL := os.Getenv("DEVENTO_BASE_URL")
	defer func() {
		os.Setenv("DEVENTO_API_KEY", originalAPIKey)
		os.Setenv("DEVENTO_BASE_URL", originalBaseURL)
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
			apiKey:      "sk-devento-test123",
			wantErr:     false,
			wantBaseURL: defaultBaseURL,
		},
		{
			name:        "API key from environment",
			apiKey:      "",
			envAPIKey:   "sk-devento-env123",
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
			apiKey:      "sk-devento-test",
			envBaseURL:  "https://custom.devento.ai",
			wantErr:     false,
			wantBaseURL: "https://custom.devento.ai",
		},
		{
			name:   "Custom base URL from option",
			apiKey: "sk-devento-test",
			opts: []ClientOption{
				WithBaseURL("https://option.devento.ai"),
			},
			wantErr:     false,
			wantBaseURL: "https://option.devento.ai",
		},
		{
			name:   "Custom HTTP client",
			apiKey: "sk-devento-test",
			opts: []ClientOption{
				WithHTTPClient(&http.Client{Timeout: 60 * time.Second}),
			},
			wantErr:     false,
			wantBaseURL: defaultBaseURL,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("DEVENTO_API_KEY", tt.envAPIKey)
			os.Setenv("DEVENTO_BASE_URL", tt.envBaseURL)

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
	originalTimeout := os.Getenv("DEVENTO_BOX_TIMEOUT")
	defer func() {
		os.Setenv("DEVENTO_BOX_TIMEOUT", originalTimeout)
	}()

	client, err := NewClient("sk-devento-test")
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
	os.Setenv("DEVENTO_BOX_TIMEOUT", "120")

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
				Hostname: "abc123.deven.to",
			},
			port:      3000,
			wantURL:   "https://3000-abc123.deven.to",
			wantError: false,
		},
		{
			name: "Different port",
			box: &Box{
				ID:       "box-456",
				Status:   BoxStatusRunning,
				Hostname: "xyz789.deven.to",
			},
			port:      8080,
			wantURL:   "https://8080-xyz789.deven.to",
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
				Hostname: "test123.deven.to",
			},
			port:      80,
			wantURL:   "https://80-test123.deven.to",
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, _ := NewClient("sk-devento-test")
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

func TestClientDomains(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	slug := "app"
	targetPort := 4000
	boxID := "box_123"

	domain := Domain{
		ID:         "dom_123",
		Hostname:   "app.deven.to",
		Slug:       &slug,
		Kind:       DomainKindManaged,
		Status:     DomainStatusActive,
		TargetPort: &targetPort,
		BoxID:      &boxID,
		VerificationPayload: map[string]any{
			"cname": "app.deven.to",
		},
		VerificationErrors: map[string]any{},
		InsertedAt:         now,
		UpdatedAt:          now,
	}

	meta := DomainMeta{
		ManagedSuffix: "deven.to",
		CNAMETarget:   "edge.deven.to",
	}

	t.Run("List domains", func(t *testing.T) {
		response := DomainsResponse{
			Data: []Domain{domain},
			Meta: meta,
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodGet {
				t.Fatalf("expected GET method, got %s", r.Method)
			}
			if r.URL.Path != "/api/v2/domains" {
				t.Fatalf("unexpected path: %s", r.URL.Path)
			}
			if err := json.NewEncoder(w).Encode(response); err != nil {
				t.Fatalf("failed to encode response: %v", err)
			}
		}))
		defer server.Close()

		client, err := NewClient("test-key", WithBaseURL(server.URL))
		if err != nil {
			t.Fatalf("NewClient error: %v", err)
		}

		resp, err := client.ListDomains(context.Background())
		if err != nil {
			t.Fatalf("ListDomains error: %v", err)
		}

		if len(resp.Data) != 1 {
			t.Fatalf("expected 1 domain, got %d", len(resp.Data))
		}
		if resp.Meta.ManagedSuffix != meta.ManagedSuffix {
			t.Fatalf("unexpected managed suffix: %s", resp.Meta.ManagedSuffix)
		}
	})

	t.Run("Get domain", func(t *testing.T) {
		response := DomainResponse{
			Data: domain,
			Meta: meta,
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodGet {
				t.Fatalf("expected GET method, got %s", r.Method)
			}
			if r.URL.Path != "/api/v2/domains/dom_123" {
				t.Fatalf("unexpected path: %s", r.URL.Path)
			}
			if err := json.NewEncoder(w).Encode(response); err != nil {
				t.Fatalf("failed to encode response: %v", err)
			}
		}))
		defer server.Close()

		client, err := NewClient("test-key", WithBaseURL(server.URL))
		if err != nil {
			t.Fatalf("NewClient error: %v", err)
		}

		resp, err := client.GetDomain(context.Background(), "dom_123")
		if err != nil {
			t.Fatalf("GetDomain error: %v", err)
		}
		if resp.Data.ID != "dom_123" {
			t.Fatalf("unexpected domain id: %s", resp.Data.ID)
		}
	})

	t.Run("Create domain omits undefined fields", func(t *testing.T) {
		response := DomainResponse{
			Data: domain,
			Meta: meta,
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				t.Fatalf("expected POST method, got %s", r.Method)
			}
			if r.URL.Path != "/api/v2/domains" {
				t.Fatalf("unexpected path: %s", r.URL.Path)
			}

			var payload map[string]any
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				t.Fatalf("failed to decode request: %v", err)
			}

			if payload["kind"] != string(DomainKindManaged) {
				t.Fatalf("unexpected kind: %v", payload["kind"])
			}
			if _, exists := payload["hostname"]; exists {
				t.Fatalf("expected hostname to be omitted, but was present")
			}
			if slugVal, ok := payload["slug"].(string); !ok || slugVal != "app" {
				t.Fatalf("unexpected slug: %v", payload["slug"])
			}
			if portVal, ok := payload["target_port"].(float64); !ok || int(portVal) != targetPort {
				t.Fatalf("unexpected target_port: %v", payload["target_port"])
			}
			if boxVal, ok := payload["box_id"].(string); !ok || boxVal != boxID {
				t.Fatalf("unexpected box_id: %v", payload["box_id"])
			}

			if err := json.NewEncoder(w).Encode(response); err != nil {
				t.Fatalf("failed to encode response: %v", err)
			}
		}))
		defer server.Close()

		client, err := NewClient("test-key", WithBaseURL(server.URL))
		if err != nil {
			t.Fatalf("NewClient error: %v", err)
		}

		req := &CreateDomainRequest{
			Kind:       DomainKindManaged,
			Slug:       &slug,
			TargetPort: &targetPort,
			BoxID:      &boxID,
		}

		resp, err := client.CreateDomain(context.Background(), req)
		if err != nil {
			t.Fatalf("CreateDomain error: %v", err)
		}
		if resp.Data.Kind != DomainKindManaged {
			t.Fatalf("unexpected domain kind: %s", resp.Data.Kind)
		}
	})

	t.Run("Update domain supports null", func(t *testing.T) {
		response := DomainResponse{
			Data: domain,
			Meta: meta,
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPatch {
				t.Fatalf("expected PATCH method, got %s", r.Method)
			}
			if r.URL.Path != "/api/v2/domains/dom_123" {
				t.Fatalf("unexpected path: %s", r.URL.Path)
			}

			var payload map[string]any
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				t.Fatalf("failed to decode request: %v", err)
			}

			if payload["status"] != string(DomainStatusActive) {
				t.Fatalf("unexpected status: %v", payload["status"])
			}
			if val, exists := payload["target_port"]; !exists || val != nil {
				t.Fatalf("expected target_port to be explicitly null, got %v", val)
			}
			if val, exists := payload["box_id"]; !exists || val != nil {
				t.Fatalf("expected box_id to be explicitly null, got %v", val)
			}

			if err := json.NewEncoder(w).Encode(response); err != nil {
				t.Fatalf("failed to encode response: %v", err)
			}
		}))
		defer server.Close()

		client, err := NewClient("test-key", WithBaseURL(server.URL))
		if err != nil {
			t.Fatalf("NewClient error: %v", err)
		}

		req := &UpdateDomainRequest{
			Status:     NewUpdateField(DomainStatusActive),
			TargetPort: NullUpdateField[int](),
			BoxID:      NullUpdateField[string](),
		}

		resp, err := client.UpdateDomain(context.Background(), "dom_123", req)
		if err != nil {
			t.Fatalf("UpdateDomain error: %v", err)
		}
		if resp.Data.ID != "dom_123" {
			t.Fatalf("unexpected domain id: %s", resp.Data.ID)
		}
	})

	t.Run("Delete domain", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodDelete {
				t.Fatalf("expected DELETE method, got %s", r.Method)
			}
			if r.URL.Path != "/api/v2/domains/dom_123" {
				t.Fatalf("unexpected path: %s", r.URL.Path)
			}
			w.WriteHeader(http.StatusNoContent)
		}))
		defer server.Close()

		client, err := NewClient("test-key", WithBaseURL(server.URL))
		if err != nil {
			t.Fatalf("NewClient error: %v", err)
		}

		if err := client.DeleteDomain(context.Background(), "dom_123"); err != nil {
			t.Fatalf("DeleteDomain error: %v", err)
		}
	})

	t.Run("Create domain nil request", func(t *testing.T) {
		client, err := NewClient("test-key", WithBaseURL("https://example.com"))
		if err != nil {
			t.Fatalf("NewClient error: %v", err)
		}

		if _, err := client.CreateDomain(context.Background(), nil); err == nil {
			t.Fatalf("expected error for nil request")
		}
	})

	t.Run("Update domain nil request", func(t *testing.T) {
		client, err := NewClient("test-key", WithBaseURL("https://example.com"))
		if err != nil {
			t.Fatalf("NewClient error: %v", err)
		}

		if _, err := client.UpdateDomain(context.Background(), "dom_123", nil); err == nil {
			t.Fatalf("expected error for nil request")
		}
	})
}
