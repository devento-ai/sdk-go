package tavor

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type BoxHandle struct {
	client     *Client
	box        *Box
	lastStdout string
	lastStderr string
}

func newBoxHandle(client *Client, box *Box) *BoxHandle {
	return &BoxHandle{
		client: client,
		box:    box,
	}
}

func (h *BoxHandle) ID() string {
	return h.box.ID
}

func (h *BoxHandle) Status() BoxStatus {
	return h.box.Status
}

func (h *BoxHandle) Metadata() map[string]string {
	return h.box.Metadata
}

func (h *BoxHandle) Refresh(ctx context.Context) error {
	var resp getBoxResponse
	err := h.client.doRequest(ctx, "GET", "/api/v2/boxes/"+h.box.ID, nil, &resp)
	if err != nil {
		return err
	}
	h.box = &resp.Data
	return nil
}

func (h *BoxHandle) WaitUntilReady(ctx context.Context) error {
	timeout := 60 * time.Second
	pollInterval := 1 * time.Second

	if envTimeout := os.Getenv("TAVOR_BOX_TIMEOUT"); envTimeout != "" {
		if t, err := strconv.Atoi(envTimeout); err == nil {
			timeout = time.Duration(t) * time.Second
		}
	}

	deadline := time.Now().Add(timeout)

	for {
		if err := h.Refresh(ctx); err != nil {
			return err
		}

		switch h.box.Status {
		case BoxStatusRunning:
			return nil
		case BoxStatusFailed, BoxStatusTerminated:
			return fmt.Errorf("box %s failed to start: %s", h.box.ID, h.box.Details)
		}

		if time.Now().After(deadline) {
			return NewBoxTimeoutError(h.box.ID, int(timeout.Seconds()))
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(pollInterval):
			// Continue polling
		}
	}
}

func (h *BoxHandle) Run(ctx context.Context, command string, opts *CommandOptions) (*CommandResult, error) {
	if opts == nil {
		opts = &CommandOptions{}
	}

	if opts.Timeout == 0 {
		opts.Timeout = 300000 // Default to 5 minutes
	}

	if opts.PollInterval == 0 {
		opts.PollInterval = 1000 // Default to 1 second
	}

	useStreaming := opts.OnStdout != nil || opts.OnStderr != nil

	if useStreaming {
		return h.runWithStreaming(ctx, command, opts)
	}

	req := queueCommandRequest{Command: command, Stream: false}
	var cmdResp queueCommandResponse
	err := h.client.doRequest(ctx, "POST", fmt.Sprintf("/api/v2/boxes/%s", h.box.ID), req, &cmdResp)
	if err != nil {
		return nil, err
	}

	commandID := cmdResp.ID
	h.client.logger.Debug("queued command", "commandID", commandID, "command", command)

	deadline := time.Now().Add(time.Duration(opts.Timeout) * time.Millisecond)
	pollInterval := time.Duration(opts.PollInterval) * time.Millisecond

	var cmd *Command
	for {
		var statusResp getCommandResponse
		err := h.client.doRequest(ctx, "GET", fmt.Sprintf("/api/v2/boxes/%s/commands/%s", h.box.ID, commandID), nil, &statusResp)
		if err != nil {
			return nil, err
		}

		cmd = (*Command)(&statusResp)

		switch cmd.Status {
		case CommandStatusDone, CommandStatusFailed, CommandStatusError:
			exitCode := 0
			if cmd.ExitCode != nil {
				exitCode = *cmd.ExitCode
			}

			return &CommandResult{
				ID:       cmd.ID,
				BoxID:    cmd.BoxID,
				Cmd:      cmd.Cmd,
				Status:   cmd.Status,
				Stdout:   cmd.Stdout,
				Stderr:   cmd.Stderr,
				ExitCode: exitCode,
			}, nil
		}

		if time.Now().After(deadline) {
			return nil, NewCommandTimeoutError(cmd.ID, opts.Timeout)
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(pollInterval):
			// Continue polling
		}
	}
}

func (h *BoxHandle) runWithStreaming(ctx context.Context, command string, opts *CommandOptions) (*CommandResult, error) {
	req := queueCommandRequest{Command: command, Stream: true}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/api/v2/boxes/%s", h.client.baseURL, h.box.ID)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, strings.NewReader(string(body)))
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-API-Key", h.client.apiKey)

	resp, err := h.client.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		var errResp errorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
			return nil, fmt.Errorf("request failed with status %d", resp.StatusCode)
		}
		return nil, fmt.Errorf("%s", errResp.Error)
	}

	events := ParseSSE(resp.Body)

	var commandID string
	var status CommandStatus = CommandStatusQueued
	var stdout, stderr string
	var exitCode int

	deadline := time.Now().Add(time.Duration(opts.Timeout) * time.Millisecond)

	for event := range events {
		if time.Now().After(deadline) {
			return nil, NewCommandTimeoutError(commandID, opts.Timeout)
		}

		switch event.Event {
		case "start":
			var data SSEStartData
			if err := ParseSSEData(event, &data); err == nil {
				commandID = data.CommandID
			}

		case "output":
			var data map[string]any
			if err := ParseSSEData(event, &data); err == nil {
				if s, ok := data["stdout"].(string); ok && s != "" {
					stdout += s
					if opts.OnStdout != nil {
						lines := strings.Split(s, "\n")
						for i, line := range lines {
							if i < len(lines)-1 || line != "" {
								opts.OnStdout(line)
							}
						}
					}
				}

				if s, ok := data["stderr"].(string); ok && s != "" {
					stderr += s
					if opts.OnStderr != nil {
						lines := strings.Split(s, "\n")
						for i, line := range lines {
							if i < len(lines)-1 || line != "" {
								opts.OnStderr(line)
							}
						}
					}
				}
			}

		case "status":
			var data map[string]any
			if err := ParseSSEData(event, &data); err == nil {
				if s, ok := data["status"].(string); ok {
					status = CommandStatus(s)
				}
				if ec, ok := data["exit_code"].(float64); ok {
					exitCode = int(ec)
				}
			}

		case "end":
			var data map[string]any
			if err := ParseSSEData(event, &data); err == nil {
				if s, ok := data["status"].(string); ok {
					if s == "error" {
						status = CommandStatusError
					} else if s == "timeout" {
						return nil, NewCommandTimeoutError(commandID, opts.Timeout)
					}
				}
			}

			return &CommandResult{
				ID:       commandID,
				BoxID:    h.box.ID,
				Cmd:      command,
				Status:   status,
				Stdout:   stdout,
				Stderr:   stderr,
				ExitCode: exitCode,
			}, nil

		case "error":
			var data map[string]any
			if err := ParseSSEData(event, &data); err == nil {
				if errMsg, ok := data["error"].(string); ok {
					return nil, fmt.Errorf("command error: %s", errMsg)
				}
			}
			return nil, fmt.Errorf("command error")

		case "timeout":
			return nil, NewCommandTimeoutError(commandID, opts.Timeout)
		}
	}

	// Stream ended without proper completion
	return &CommandResult{
		ID:       commandID,
		BoxID:    h.box.ID,
		Cmd:      command,
		Status:   status,
		Stdout:   stdout,
		Stderr:   stderr,
		ExitCode: exitCode,
	}, nil
}

func (h *BoxHandle) streamOutput(cmd *Command, opts *CommandOptions) {
	if opts.OnStdout != nil && len(cmd.Stdout) > len(h.lastStdout) {
		newOutput := cmd.Stdout[len(h.lastStdout):]
		h.lastStdout = cmd.Stdout

		// Split into lines and call callback
		scanner := bufio.NewScanner(strings.NewReader(newOutput))
		for scanner.Scan() {
			opts.OnStdout(scanner.Text())
		}
	}

	if opts.OnStderr != nil && len(cmd.Stderr) > len(h.lastStderr) {
		newOutput := cmd.Stderr[len(h.lastStderr):]
		h.lastStderr = cmd.Stderr

		// Split into lines and call callback
		scanner := bufio.NewScanner(strings.NewReader(newOutput))
		for scanner.Scan() {
			opts.OnStderr(scanner.Text())
		}
	}
}

func (h *BoxHandle) Stop(ctx context.Context) error {
	return h.client.doRequest(ctx, "DELETE", "/api/v2/boxes/"+h.box.ID, nil, nil)
}

// Close is an alias for Stop for consistency with other SDKs
func (h *BoxHandle) Close(ctx context.Context) error {
	return h.Stop(ctx)
}

// GetPublicURL returns the public web URL for accessing a specific port on the box
func (h *BoxHandle) GetPublicURL(port int) (string, error) {
	if h.box.Hostname == "" {
		return "", fmt.Errorf("box does not have a hostname. Ensure the box is created and running")
	}
	return fmt.Sprintf("https://%d-%s", port, h.box.Hostname), nil
}

// ExposePort exposes a port from inside the sandbox to a random external port.
// This allows external access to services running inside the sandbox.
//
// targetPort is the port number inside the sandbox to expose.
// Returns an ExposedPort containing the proxy_port (external), target_port, and expires_at.
// Returns an error if the box is not in a running state or if no ports are available.
func (h *BoxHandle) ExposePort(ctx context.Context, targetPort int) (*ExposedPort, error) {
	req := exposePortRequest{Port: targetPort}
	var resp exposePortResponse

	err := h.client.doRequest(ctx, "POST", fmt.Sprintf("/api/v2/boxes/%s/expose_port", h.box.ID), req, &resp)
	if err != nil {
		return nil, err
	}

	return &resp.Data, nil
}
