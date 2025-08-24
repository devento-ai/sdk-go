package devento

import (
	"time"
)

type BoxStatus string

const (
	BoxStatusQueued     BoxStatus = "queued"
	BoxStatusStarting   BoxStatus = "starting"
	BoxStatusRunning    BoxStatus = "running"
	BoxStatusPaused     BoxStatus = "paused"
	BoxStatusStopping   BoxStatus = "stopping"
	BoxStatusStopped    BoxStatus = "stopped"
	BoxStatusFailed     BoxStatus = "failed"
	BoxStatusTerminated BoxStatus = "terminated"
)

type CommandStatus string

const (
	CommandStatusQueued  CommandStatus = "queued"
	CommandStatusRunning CommandStatus = "running"
	CommandStatusDone    CommandStatus = "done"
	CommandStatusFailed  CommandStatus = "failed"
	CommandStatusError   CommandStatus = "error"
)

type Box struct {
	ID           string            `json:"id"`
	Status       BoxStatus         `json:"status"`
	Metadata     map[string]string `json:"metadata,omitempty"`
	StartedAt    *time.Time        `json:"started_at,omitempty"`
	TerminatedAt *time.Time        `json:"terminated_at,omitempty"`
	Details      string            `json:"details,omitempty"`
	InsertedAt   time.Time         `json:"created_at"`
	Hostname     string            `json:"hostname,omitempty"`
}

type Command struct {
	ID        string        `json:"id"`
	BoxID     string        `json:"box_id"`
	Cmd       string        `json:"cmd"`
	Status    CommandStatus `json:"status"`
	Stdout    string        `json:"stdout,omitempty"`
	Stderr    string        `json:"stderr,omitempty"`
	ExitCode  *int          `json:"exit_code,omitempty"`
	CreatedAt time.Time     `json:"created_at"`
	UpdatedAt time.Time     `json:"updated_at"`
}

type CommandResult struct {
	ID       string        `json:"id"`
	BoxID    string        `json:"box_id"`
	Cmd      string        `json:"cmd"`
	Status   CommandStatus `json:"status"`
	Stdout   string        `json:"stdout"`
	Stderr   string        `json:"stderr"`
	ExitCode int           `json:"exit_code"`
}

type BoxConfig struct {
	CPU      int               `json:"cpu,omitempty"`     // Number of CPU cores
	MibRAM   int               `json:"mib_ram,omitempty"` // RAM in MiB
	Timeout  int               `json:"timeout,omitempty"` // seconds
	Metadata map[string]string `json:"metadata,omitempty"`
}

type CommandOptions struct {
	Timeout      int               `json:"timeout,omitempty"`       // milliseconds
	PollInterval int               `json:"poll_interval,omitempty"` // milliseconds
	OnStdout     func(line string) `json:"-"`
	OnStderr     func(line string) `json:"-"`
}

type Organization struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	Credits    float64   `json:"credits"`
	InsertedAt time.Time `json:"inserted_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// API request/response types

type createBoxRequest struct {
	CPU      int               `json:"cpu,omitempty"`
	MibRAM   int               `json:"mib_ram,omitempty"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

type createBoxResponse struct {
	ID string `json:"id"`
}

type listBoxesResponse struct {
	Data []Box `json:"data"`
}

type getBoxResponse struct {
	Data Box `json:"data"`
}

type queueCommandRequest struct {
	Command   string `json:"command"`
	Stream    bool   `json:"stream,omitempty"`
	TimeoutMs *int   `json:"timeout_ms,omitempty"`
}

type queueCommandResponse struct {
	ID string `json:"id"`
}

type getCommandResponse Command

type errorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Code    string `json:"code,omitempty"`
}

type ExposedPort struct {
	ProxyPort  int       `json:"proxy_port"`
	TargetPort int       `json:"target_port"`
	ExpiresAt  time.Time `json:"expires_at"`
}

type exposePortRequest struct {
	Port int `json:"port"`
}

type exposePortResponse struct {
	Data ExposedPort `json:"data"`
}

type SnapshotStatus string

const (
	SnapshotStatusCreating  SnapshotStatus = "creating"
	SnapshotStatusReady     SnapshotStatus = "ready"
	SnapshotStatusRestoring SnapshotStatus = "restoring"
	SnapshotStatusDeleted   SnapshotStatus = "deleted"
	SnapshotStatusError     SnapshotStatus = "error"
)

type Snapshot struct {
	ID             string         `json:"id"`
	BoxID          string         `json:"box_id"`
	SnapshotType   string         `json:"snapshot_type"`
	Status         SnapshotStatus `json:"status"`
	Label          string         `json:"label,omitempty"`
	SizeBytes      *int64         `json:"size_bytes,omitempty"`
	ChecksumSHA256 string         `json:"checksum_sha256,omitempty"`
	CreatedAt      time.Time      `json:"created_at"`
	OrchestratorID string         `json:"orchestrator_id"`
}

type listSnapshotsResponse struct {
	Data []Snapshot `json:"data"`
}

type getSnapshotResponse struct {
	Data Snapshot `json:"data"`
}
