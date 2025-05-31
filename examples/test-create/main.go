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

	// Create a box
	fmt.Println("Creating box...")
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
