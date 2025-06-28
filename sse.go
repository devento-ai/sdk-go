package tavor

import (
	"bufio"
	"encoding/json"
	"io"
	"strings"
)

type SSEEvent struct {
	Event string
	Data  string
}

type SSEOutputData struct {
	Stdout string `json:"stdout,omitempty"`
	Stderr string `json:"stderr,omitempty"`
}

type SSEStatusData struct {
	Status   string `json:"status"`
	ExitCode *int   `json:"exit_code,omitempty"`
}

type SSEEndData struct {
	Status string `json:"status"`
}

type SSEErrorData struct {
	Error string `json:"error"`
}

type SSEStartData struct {
	CommandID string `json:"command_id"`
	Status    string `json:"status"`
}

func ParseSSE(reader io.Reader) <-chan SSEEvent {
	events := make(chan SSEEvent)

	go func() {
		defer close(events)
		scanner := bufio.NewScanner(reader)

		var event, data string

		for scanner.Scan() {
			line := scanner.Text()

			if line == "" {
				if event != "" && data != "" {
					events <- SSEEvent{
						Event: event,
						Data:  data,
					}
				}
				event = ""
				data = ""
				continue
			}

			if after, found := strings.CutPrefix(line, "event: "); found {
				event = after
			} else if after, found := strings.CutPrefix(line, "data: "); found {
				data = after
			}
		}
	}()

	return events
}

func ParseSSEData(event SSEEvent, v any) error {
	return json.Unmarshal([]byte(event.Data), v)
}
