package main

import (
	"context"
	"fmt"
	"log"

	tavor "github.com/tavor-dev/sdk-go"
)

func main() {
	// Create client with debug enabled to see the requests
	client, err := tavor.NewClient("", tavor.WithDebug(true))
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	fmt.Println("Testing command execution with new endpoint format...")

	err = client.WithSandbox(ctx, func(ctx context.Context, box *tavor.BoxHandle) error {
		// Test a simple command
		fmt.Println("\nRunning echo command...")
		result, err := box.Run(ctx, "echo 'Hello from fixed Go SDK!'", nil)
		if err != nil {
			return fmt.Errorf("echo command failed: %w", err)
		}

		fmt.Printf("Output: %s", result.Stdout)
		fmt.Printf("Exit code: %d\n", result.ExitCode)

		// Test command with streaming
		fmt.Println("\nRunning command with streaming output...")
		opts := &tavor.CommandOptions{
			OnStdout: func(line string) {
				fmt.Printf("[STREAM] %s\n", line)
			},
		}

		_, err = box.Run(ctx, "for i in 1 2 3; do echo Line $i; done", opts)
		if err != nil {
			return fmt.Errorf("streaming command failed: %w", err)
		}

		return nil
	}, nil)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("\nAll tests passed!")
}
