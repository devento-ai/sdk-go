package main

import (
	"context"
	"fmt"
	"log"

	devento "github.com/devento-ai/sdk-go"
)

func main() {
	// Create client (uses DEVENTO_API_KEY env var if not provided)
	client, err := devento.NewClient("", devento.WithDebug(true))
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	// Use WithSandbox for automatic cleanup
	err = client.WithSandbox(ctx, func(ctx context.Context, box *devento.BoxHandle) error {
		result, err := box.Run(ctx, "echo 'Hello from Devento!'", nil)
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
