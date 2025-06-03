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

	// Create box with custom configuration
	// Since we're running Python scripts, we might want specific resources
	config := &tavor.BoxConfig{
		CPU:    1,   // 1 CPU cores
		MibRAM: 256, // 256 MiB RAM for Python runtime
		Metadata: map[string]string{
			"project": "demo",
			"type":    "python",
		},
	}

	box, err := client.CreateBox(ctx, config)
	if err != nil {
		log.Fatal(err)
	}
	defer box.Stop(ctx) // Always clean up

	// Wait for box to be ready
	if err := box.WaitUntilReady(ctx); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Box %s is ready!\n", box.ID())

	// Create a Python script
	script := `
print("Hello from Python!")
import sys
print(f"Python version: {sys.version}")

# Calculate something
numbers = [1, 2, 3, 4, 5]
total = sum(numbers)
print(f"Sum of {numbers} = {total}")
`

	result, err := box.Run(ctx, fmt.Sprintf("echo '%s' > script.py", script), nil)
	if err != nil {
		log.Fatal(err)
	}

	// Run the Python script
	result, err = box.Run(ctx, "python script.py", nil)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Python output:")
	fmt.Println(result.Stdout)
	fmt.Println(result.Stderr)
}
