package devento

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestBoxHandle_ListSnapshots(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Expected GET request, got %s", r.Method)
		}
		if r.URL.Path != "/api/v2/boxes/test-box-id/snapshots" {
			t.Errorf("Expected path /api/v2/boxes/test-box-id/snapshots, got %s", r.URL.Path)
		}

		response := listSnapshotsResponse{
			Data: []Snapshot{
				{
					ID:             "snap-1",
					BoxID:          "test-box-id",
					SnapshotType:   "disk",
					Status:         SnapshotStatusReady,
					Label:          "backup-1",
					CreatedAt:      time.Now(),
					OrchestratorID: "orch-1",
				},
				{
					ID:             "snap-2",
					BoxID:          "test-box-id",
					SnapshotType:   "disk",
					Status:         SnapshotStatusCreating,
					CreatedAt:      time.Now(),
					OrchestratorID: "orch-1",
				},
			},
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, _ := NewClient("test-api-key", WithBaseURL(server.URL))
	box := &Box{ID: "test-box-id", Status: BoxStatusRunning}
	handle := newBoxHandle(client, box)

	snapshots, err := handle.ListSnapshots(context.Background())
	if err != nil {
		t.Fatalf("ListSnapshots failed: %v", err)
	}

	if len(snapshots) != 2 {
		t.Errorf("Expected 2 snapshots, got %d", len(snapshots))
	}

	if snapshots[0].ID != "snap-1" {
		t.Errorf("Expected first snapshot ID 'snap-1', got %s", snapshots[0].ID)
	}

	if snapshots[0].Status != SnapshotStatusReady {
		t.Errorf("Expected first snapshot status 'ready', got %s", snapshots[0].Status)
	}
}

func TestBoxHandle_GetSnapshot(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Expected GET request, got %s", r.Method)
		}
		if r.URL.Path != "/api/v2/boxes/test-box-id/snapshots/snap-1" {
			t.Errorf("Expected path /api/v2/boxes/test-box-id/snapshots/snap-1, got %s", r.URL.Path)
		}

		response := getSnapshotResponse{
			Data: Snapshot{
				ID:             "snap-1",
				BoxID:          "test-box-id",
				SnapshotType:   "disk",
				Status:         SnapshotStatusReady,
				Label:          "backup-1",
				CreatedAt:      time.Now(),
				OrchestratorID: "orch-1",
			},
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, _ := NewClient("test-api-key", WithBaseURL(server.URL))
	box := &Box{ID: "test-box-id", Status: BoxStatusRunning}
	handle := newBoxHandle(client, box)

	snapshot, err := handle.GetSnapshot(context.Background(), "snap-1")
	if err != nil {
		t.Fatalf("GetSnapshot failed: %v", err)
	}

	if snapshot.ID != "snap-1" {
		t.Errorf("Expected snapshot ID 'snap-1', got %s", snapshot.ID)
	}

	if snapshot.Status != SnapshotStatusReady {
		t.Errorf("Expected snapshot status 'ready', got %s", snapshot.Status)
	}
}

func TestBoxHandle_CreateSnapshot(t *testing.T) {
	tests := []struct {
		name        string
		label       string
		description string
		wantLabel   bool
		wantDesc    bool
	}{
		{
			name:        "With label and description",
			label:       "before-upgrade",
			description: "Snapshot before system upgrade",
			wantLabel:   true,
			wantDesc:    true,
		},
		{
			name:      "With label only",
			label:     "backup",
			wantLabel: true,
		},
		{
			name: "Without parameters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != "POST" {
					t.Errorf("Expected POST request, got %s", r.Method)
				}
				if r.URL.Path != "/api/v2/boxes/test-box-id/snapshots" {
					t.Errorf("Expected path /api/v2/boxes/test-box-id/snapshots, got %s", r.URL.Path)
				}

				var reqBody map[string]string
				if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
					t.Errorf("Failed to decode request body: %v", err)
				}

				if tt.wantLabel {
					if reqBody["label"] != tt.label {
						t.Errorf("Expected label '%s', got '%s'", tt.label, reqBody["label"])
					}
				}

				if tt.wantDesc {
					if reqBody["description"] != tt.description {
						t.Errorf("Expected description '%s', got '%s'", tt.description, reqBody["description"])
					}
				}

				response := getSnapshotResponse{
					Data: Snapshot{
						ID:             "snap-new",
						BoxID:          "test-box-id",
						SnapshotType:   "disk",
						Status:         SnapshotStatusCreating,
						Label:          tt.label,
						CreatedAt:      time.Now(),
						OrchestratorID: "orch-1",
					},
				}

				w.WriteHeader(http.StatusCreated)
				json.NewEncoder(w).Encode(response)
			}))
			defer server.Close()

			client, _ := NewClient("test-api-key", WithBaseURL(server.URL))
			box := &Box{ID: "test-box-id", Status: BoxStatusRunning}
			handle := newBoxHandle(client, box)

			snapshot, err := handle.CreateSnapshot(context.Background(), tt.label, tt.description)
			if err != nil {
				t.Fatalf("CreateSnapshot failed: %v", err)
			}

			if snapshot.ID != "snap-new" {
				t.Errorf("Expected snapshot ID 'snap-new', got %s", snapshot.ID)
			}

			if snapshot.Status != SnapshotStatusCreating {
				t.Errorf("Expected snapshot status 'creating', got %s", snapshot.Status)
			}
		})
	}
}

func TestBoxHandle_RestoreSnapshot(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		if r.URL.Path != "/api/v2/boxes/test-box-id/restore" {
			t.Errorf("Expected path /api/v2/boxes/test-box-id/restore, got %s", r.URL.Path)
		}

		var reqBody map[string]string
		if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
			t.Errorf("Failed to decode request body: %v", err)
		}

		if reqBody["snapshot_id"] != "snap-1" {
			t.Errorf("Expected snapshot_id 'snap-1', got '%s'", reqBody["snapshot_id"])
		}

		response := getSnapshotResponse{
			Data: Snapshot{
				ID:             "snap-1",
				BoxID:          "test-box-id",
				SnapshotType:   "disk",
				Status:         SnapshotStatusRestoring,
				CreatedAt:      time.Now(),
				OrchestratorID: "orch-1",
			},
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, _ := NewClient("test-api-key", WithBaseURL(server.URL))
	box := &Box{ID: "test-box-id", Status: BoxStatusRunning}
	handle := newBoxHandle(client, box)

	snapshot, err := handle.RestoreSnapshot(context.Background(), "snap-1")
	if err != nil {
		t.Fatalf("RestoreSnapshot failed: %v", err)
	}

	if snapshot.Status != SnapshotStatusRestoring {
		t.Errorf("Expected snapshot status 'restoring', got %s", snapshot.Status)
	}
}

func TestBoxHandle_DeleteSnapshot(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("Expected DELETE request, got %s", r.Method)
		}
		if r.URL.Path != "/api/v2/boxes/test-box-id/snapshots/snap-1" {
			t.Errorf("Expected path /api/v2/boxes/test-box-id/snapshots/snap-1, got %s", r.URL.Path)
		}

		response := getSnapshotResponse{
			Data: Snapshot{
				ID:             "snap-1",
				BoxID:          "test-box-id",
				SnapshotType:   "disk",
				Status:         SnapshotStatusDeleted,
				CreatedAt:      time.Now(),
				OrchestratorID: "orch-1",
			},
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, _ := NewClient("test-api-key", WithBaseURL(server.URL))
	box := &Box{ID: "test-box-id", Status: BoxStatusRunning}
	handle := newBoxHandle(client, box)

	snapshot, err := handle.DeleteSnapshot(context.Background(), "snap-1")
	if err != nil {
		t.Fatalf("DeleteSnapshot failed: %v", err)
	}

	if snapshot.Status != SnapshotStatusDeleted {
		t.Errorf("Expected snapshot status 'deleted', got %s", snapshot.Status)
	}
}

func TestBoxHandle_WaitSnapshotReady(t *testing.T) {
	tests := []struct {
		name      string
		responses []SnapshotStatus
		wantErr   bool
		errMsg    string
	}{
		{
			name:      "Becomes ready",
			responses: []SnapshotStatus{SnapshotStatusCreating, SnapshotStatusCreating, SnapshotStatusReady},
			wantErr:   false,
		},
		{
			name:      "Ends in error",
			responses: []SnapshotStatus{SnapshotStatusError},
			wantErr:   true,
			errMsg:    "ended with status: error",
		},
		{
			name:      "Gets deleted",
			responses: []SnapshotStatus{SnapshotStatusDeleted},
			wantErr:   true,
			errMsg:    "ended with status: deleted",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			callCount := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != "GET" {
					t.Errorf("Expected GET request, got %s", r.Method)
				}
				if r.URL.Path != "/api/v2/boxes/test-box-id/snapshots/snap-1" {
					t.Errorf("Expected path /api/v2/boxes/test-box-id/snapshots/snap-1, got %s", r.URL.Path)
				}

				status := tt.responses[callCount]
				if callCount < len(tt.responses)-1 {
					callCount++
				}

				response := getSnapshotResponse{
					Data: Snapshot{
						ID:             "snap-1",
						BoxID:          "test-box-id",
						SnapshotType:   "disk",
						Status:         status,
						CreatedAt:      time.Now(),
						OrchestratorID: "orch-1",
					},
				}

				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(response)
			}))
			defer server.Close()

			client, _ := NewClient("test-api-key", WithBaseURL(server.URL))
			box := &Box{ID: "test-box-id", Status: BoxStatusRunning}
			handle := newBoxHandle(client, box)

			err := handle.WaitSnapshotReady(context.Background(), "snap-1", 5*time.Second, 10*time.Millisecond)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error, got nil")
				} else if tt.errMsg != "" && err.Error() != fmt.Sprintf("snapshot snap-1 %s", tt.errMsg) {
					t.Errorf("Expected error message containing '%s', got '%s'", tt.errMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got %v", err)
				}
			}
		})
	}
}

func TestBoxHandle_WaitSnapshotReady_Timeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := getSnapshotResponse{
			Data: Snapshot{
				ID:             "snap-1",
				BoxID:          "test-box-id",
				SnapshotType:   "disk",
				Status:         SnapshotStatusCreating,
				CreatedAt:      time.Now(),
				OrchestratorID: "orch-1",
			},
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, _ := NewClient("test-api-key", WithBaseURL(server.URL))
	box := &Box{ID: "test-box-id", Status: BoxStatusRunning}
	handle := newBoxHandle(client, box)

	err := handle.WaitSnapshotReady(context.Background(), "snap-1", 50*time.Millisecond, 10*time.Millisecond)

	if err == nil {
		t.Errorf("Expected timeout error, got nil")
	}

	cmdErr, ok := err.(*CommandTimeoutError)
	if !ok {
		t.Errorf("Expected CommandTimeoutError, got %T", err)
	}

	if cmdErr.Timeout != 50 {
		t.Errorf("Expected timeout of 50ms, got %dms", cmdErr.Timeout)
	}
}
