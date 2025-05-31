package main

import (
	"context"
	"fmt"
	"log"

	tavor "github.com/tavor-dev/sdk-go"
)

func main() {
	// Create client (uses TAVOR_API_KEY env var if not provided)
	client, err := tavor.NewClient("", tavor.WithDebug(true))
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	// Use WithSandbox for automatic cleanup
	err = client.WithSandbox(ctx, func(ctx context.Context, box *tavor.BoxHandle) error {
		result, err := box.Run(ctx, "echo 'Hello from Tavor!'", nil)
		if err != nil {
			return err
		}

		fmt.Printf("Output: %s\n", result.Stdout)
		fmt.Printf("Exit code: %d\n", result.ExitCode)

		// Run multiple commands
		result, err = box.Run(ctx, "pwd", nil)
		if err != nil {
			return err
		}
		fmt.Printf("Working directory: %s\n", result.Stdout)

		return nil
	}, nil)
	if err != nil {
		log.Printf("Error: %v", err)
	}
}
