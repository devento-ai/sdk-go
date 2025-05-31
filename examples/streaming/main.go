package main

import (
	"context"
	"fmt"
	"log"

	tavor "github.com/tavor-dev/sdk-go"
)

func main() {
	client, err := tavor.NewClient("")
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	err = client.WithSandbox(ctx, func(ctx context.Context, box *tavor.BoxHandle) error {
		// Run a command with streaming output
		opts := &tavor.CommandOptions{
			OnStdout: func(line string) {
				fmt.Printf("[STDOUT] %s\n", line)
			},
			OnStderr: func(line string) {
				fmt.Printf("[STDERR] %s\n", line)
			},
		}

		// Simulate a long-running process
		script := `
for i in {1..5}; do
    echo "Progress: Step $i of 5"
    sleep 1
done
echo "Task completed!"
`

		_, err := box.Run(ctx, fmt.Sprintf("bash -c '%s'", script), opts)
		return err
	}, nil)

	if err != nil {
		log.Printf("Error: %v", err)
	}
}
