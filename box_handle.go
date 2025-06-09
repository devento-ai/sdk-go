package tavor

import (
	"bufio"
	"context"
	"fmt"
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

	req := queueCommandRequest{Command: command}
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

		if opts.OnStdout != nil || opts.OnStderr != nil {
			h.streamOutput(cmd, opts)
		}

		switch cmd.Status {
		case CommandStatusDone, CommandStatusFailed, CommandStatusError:
			exitCode := 0
			if cmd.ExitCode != nil {
				exitCode = *cmd.ExitCode
			}

			// Stream any remaining output
			if opts.OnStdout != nil || opts.OnStderr != nil {
				h.streamOutput(cmd, opts)
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
