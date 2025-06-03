package main

import (
	"context"
	"fmt"
	"log"

	tavor "github.com/tavor-dev/sdk-go"
)

func main() {
	// Create client with debug enabled to see the response
	client, err := tavor.NewClient("", tavor.WithDebug(true))
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	// Example 1: Create with default resources
	defaultExample(ctx, client)

	// Example 2: Create with explicit resources
	explicitExample(ctx, client)
}

func defaultExample(ctx context.Context, client *tavor.Client) {
	fmt.Println("\n=== Example 1: Default Resources ===")

	// Create a box with minimal resources (using defaults)
	fmt.Println("Creating box with default resources (1 CPU, 1024 MiB RAM)...")
	box, err := client.CreateBox(ctx, nil)
	if err != nil {
		log.Fatalf("Failed to create box: %v", err)
	}

	fmt.Printf("Created box with ID: %s\n", box.ID())
	fmt.Printf("Initial status: %s\n", box.Status())

	// Clean up
	defer func() {
		fmt.Println("Cleaning up box...")
		if err := box.Stop(ctx); err != nil {
			log.Printf("Failed to stop box: %v", err)
		}
	}()

	// Wait for box to be ready
	fmt.Println("Waiting for box to be ready...")
	if err := box.WaitUntilReady(ctx); err != nil {
		log.Fatalf("Box failed to become ready: %v", err)
	}

	fmt.Printf("Box is ready! Status: %s\n", box.Status())
}

func explicitExample(ctx context.Context, client *tavor.Client) {
	fmt.Println("\n=== Example 2: Explicit Resources ===")

	// Create a box with explicit resource configuration
	config := &tavor.BoxConfig{
		CPU:    2,    // 2 CPU cores
		MibRAM: 1024, // 1 GiB RAM (1024 MiB)
		Metadata: map[string]string{
			"environment": "development",
			"project":     "resource-test",
		},
	}

	fmt.Printf("Creating box with explicit resources (%d CPU, %d MiB RAM)...\n", config.CPU, config.MibRAM)
	box, err := client.CreateBox(ctx, config)
	if err != nil {
		log.Fatalf("Failed to create box: %v", err)
	}

	fmt.Printf("Created box with ID: %s\n", box.ID())
	fmt.Printf("Initial status: %s\n", box.Status())

	// Clean up
	defer func() {
		fmt.Println("Cleaning up box...")
		if err := box.Stop(ctx); err != nil {
			log.Printf("Failed to stop box: %v", err)
		}
	}()

	// Wait for box to be ready
	fmt.Println("Waiting for box to be ready...")
	if err := box.WaitUntilReady(ctx); err != nil {
		log.Fatalf("Box failed to become ready: %v", err)
	}

	fmt.Printf("Box is ready! Status: %s\n", box.Status())
}
