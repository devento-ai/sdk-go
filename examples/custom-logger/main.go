package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"

	devento "github.com/devento-ai/sdk-go"
)

func main() {
	// Example 1: Using the built-in debug mode
	fmt.Println("=== Example 1: Built-in debug mode ===")
	client1, err := devento.NewClient("", devento.WithDebug(true))
	if err != nil {
		log.Fatal(err)
	}

	// Example 2: Using a custom logger with JSON format
	fmt.Println("\n=== Example 2: Custom JSON logger ===")
	jsonLogger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	client2, err := devento.NewClient("", devento.WithLogger(jsonLogger))
	if err != nil {
		log.Fatal(err)
	}

	// Example 3: Using a custom logger with custom handler
	fmt.Println("\n=== Example 3: Custom logger with attributes ===")
	customLogger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})).With("service", "my-app", "version", "1.0.0")
	client3, err := devento.NewClient("", devento.WithLogger(customLogger))
	if err != nil {
		log.Fatal(err)
	}

	// Test with each client
	ctx := context.Background()
	for i, client := range []*devento.Client{client1, client2, client3} {
		fmt.Printf("\n--- Testing with client %d ---\n", i+1)
		err = client.WithSandbox(ctx, func(ctx context.Context, box *devento.BoxHandle) error {
			result, err := box.Run(ctx, "echo 'Hello from custom logger example'", nil)
			if err != nil {
				return err
			}
			fmt.Printf("Output: %s", result.Stdout)
			return nil
		}, nil)
		if err != nil {
			log.Printf("Client %d error: %v", i+1, err)
		}
	}
}
