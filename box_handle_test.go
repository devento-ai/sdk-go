package devento

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestBoxHandle_ExposePort(t *testing.T) {
	tests := []struct {
		name       string
		targetPort int
		response   exposePortResponse
		statusCode int
		wantErr    bool
		wantProxy  int
		wantTarget int
	}{
		{
			name:       "Successful port exposure",
			targetPort: 3000,
			response: exposePortResponse{
				Data: ExposedPort{
					ProxyPort:  12345,
					TargetPort: 3000,
					ExpiresAt:  time.Now().Add(1 * time.Hour),
				},
			},
			statusCode: http.StatusCreated,
			wantErr:    false,
			wantProxy:  12345,
			wantTarget: 3000,
		},
		{
			name:       "Different port",
			targetPort: 8080,
			response: exposePortResponse{
				Data: ExposedPort{
					ProxyPort:  54321,
					TargetPort: 8080,
					ExpiresAt:  time.Now().Add(1 * time.Hour),
				},
			},
			statusCode: http.StatusCreated,
			wantErr:    false,
			wantProxy:  54321,
			wantTarget: 8080,
		},
		{
			name:       "Box not running",
			targetPort: 3000,
			statusCode: http.StatusConflict,
			wantErr:    true,
		},
		{
			name:       "No ports available",
			targetPort: 3000,
			statusCode: http.StatusServiceUnavailable,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify request
				if r.Method != "POST" {
					t.Errorf("Expected POST request, got %s", r.Method)
				}
				if r.URL.Path != "/api/v2/boxes/test-box-id/expose_port" {
					t.Errorf("Expected path /api/v2/boxes/test-box-id/expose_port, got %s", r.URL.Path)
				}

				// Check request body
				var reqBody exposePortRequest
				if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
					t.Errorf("Failed to decode request body: %v", err)
				}
				if reqBody.Port != tt.targetPort {
					t.Errorf("Expected port %d, got %d", tt.targetPort, reqBody.Port)
				}

				// Send response
				w.WriteHeader(tt.statusCode)
				if tt.statusCode == http.StatusCreated {
					json.NewEncoder(w).Encode(tt.response)
				} else {
					json.NewEncoder(w).Encode(errorResponse{
						Error: "Error message",
					})
				}
			}))
			defer server.Close()

			// Create client and box handle
			client, _ := NewClient("test-api-key", WithBaseURL(server.URL))
			box := &Box{ID: "test-box-id", Status: BoxStatusRunning}
			handle := newBoxHandle(client, box)

			// Test ExposePort
			ctx := context.Background()
			result, err := handle.ExposePort(ctx, tt.targetPort)

			// Check results
			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if result.ProxyPort != tt.wantProxy {
					t.Errorf("Expected proxy port %d, got %d", tt.wantProxy, result.ProxyPort)
				}
				if result.TargetPort != tt.wantTarget {
					t.Errorf("Expected target port %d, got %d", tt.wantTarget, result.TargetPort)
				}
			}
		})
	}
}

func TestBoxHandle_Pause(t *testing.T) {
	tests := []struct {
		name          string
		initialStatus BoxStatus
		pausedStatus  BoxStatus
		statusCode    int
		wantErr       bool
	}{
		{
			name:          "Successful pause",
			initialStatus: BoxStatusRunning,
			pausedStatus:  BoxStatusStopped,
			statusCode:    http.StatusOK,
			wantErr:       false,
		},
		{
			name:          "Failed pause",
			initialStatus: BoxStatusRunning,
			statusCode:    http.StatusUnprocessableEntity,
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			requestCount := 0
			// Create test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requestCount++

				if requestCount == 1 {
					// First request should be the pause request
					if r.Method != "POST" {
						t.Errorf("Expected POST request, got %s", r.Method)
					}
					if r.URL.Path != "/api/v2/boxes/test-box-id/pause" {
						t.Errorf("Expected path /api/v2/boxes/test-box-id/pause, got %s", r.URL.Path)
					}

					w.WriteHeader(tt.statusCode)
					if tt.statusCode != http.StatusOK {
						json.NewEncoder(w).Encode(errorResponse{
							Error: "Cannot pause box",
						})
					}
				} else if requestCount == 2 && !tt.wantErr {
					// Second request should be the refresh request (GET box)
					if r.Method != "GET" {
						t.Errorf("Expected GET request for refresh, got %s", r.Method)
					}
					if r.URL.Path != "/api/v2/boxes/test-box-id" {
						t.Errorf("Expected path /api/v2/boxes/test-box-id, got %s", r.URL.Path)
					}

					w.WriteHeader(http.StatusOK)
					json.NewEncoder(w).Encode(getBoxResponse{
						Data: Box{
							ID:     "test-box-id",
							Status: tt.pausedStatus,
						},
					})
				}
			}))
			defer server.Close()

			// Create client and box handle
			client, _ := NewClient("test-api-key", WithBaseURL(server.URL))
			box := &Box{ID: "test-box-id", Status: tt.initialStatus}
			handle := newBoxHandle(client, box)

			// Test Pause
			ctx := context.Background()
			err := handle.Pause(ctx)

			// Check results
			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				// Verify that refresh was called
				if requestCount != 2 {
					t.Errorf("Expected 2 requests (pause + refresh), got %d", requestCount)
				}
			}
		})
	}
}

func TestBoxHandle_Resume(t *testing.T) {
	tests := []struct {
		name          string
		initialStatus BoxStatus
		resumedStatus BoxStatus
		statusCode    int
		wantErr       bool
	}{
		{
			name:          "Successful resume",
			initialStatus: BoxStatusStopped,
			resumedStatus: BoxStatusRunning,
			statusCode:    http.StatusOK,
			wantErr:       false,
		},
		{
			name:          "Failed resume",
			initialStatus: BoxStatusStopped,
			statusCode:    http.StatusUnprocessableEntity,
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			requestCount := 0
			// Create test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requestCount++

				if requestCount == 1 {
					// First request should be the resume request
					if r.Method != "POST" {
						t.Errorf("Expected POST request, got %s", r.Method)
					}
					if r.URL.Path != "/api/v2/boxes/test-box-id/resume" {
						t.Errorf("Expected path /api/v2/boxes/test-box-id/resume, got %s", r.URL.Path)
					}

					w.WriteHeader(tt.statusCode)
					if tt.statusCode != http.StatusOK {
						json.NewEncoder(w).Encode(errorResponse{
							Error: "Cannot resume box",
						})
					}
				} else if requestCount == 2 && !tt.wantErr {
					// Second request should be the refresh request (GET box)
					if r.Method != "GET" {
						t.Errorf("Expected GET request for refresh, got %s", r.Method)
					}
					if r.URL.Path != "/api/v2/boxes/test-box-id" {
						t.Errorf("Expected path /api/v2/boxes/test-box-id, got %s", r.URL.Path)
					}

					w.WriteHeader(http.StatusOK)
					json.NewEncoder(w).Encode(getBoxResponse{
						Data: Box{
							ID:     "test-box-id",
							Status: tt.resumedStatus,
						},
					})
				}
			}))
			defer server.Close()

			// Create client and box handle
			client, _ := NewClient("test-api-key", WithBaseURL(server.URL))
			box := &Box{ID: "test-box-id", Status: tt.initialStatus}
			handle := newBoxHandle(client, box)

			// Test Resume
			ctx := context.Background()
			err := handle.Resume(ctx)

			// Check results
			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				// Verify that refresh was called
				if requestCount != 2 {
					t.Errorf("Expected 2 requests (resume + refresh), got %d", requestCount)
				}
			}
		})
	}
}
