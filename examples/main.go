package main

import (
	"context"
	"fmt"
	"log"
	"time"

	tavor "github.com/tavor-dev/sdk-go"
)

func main() {
	// Example 1: Basic usage with automatic cleanup
	basicExample()

	// Example 2: Manual box management
	manualExample()

	// Example 3: Streaming output
	streamingExample()

	// Example 4: Concurrent commands
	concurrentExample()

	// Example 5: Error handling
	errorHandlingExample()
}

func basicExample() {
	fmt.Println("=== Basic Example: Hello World ===")

	// Create client (uses TAVOR_API_KEY env var if not provided)
	client, err := tavor.NewClient("", tavor.WithDebug(true))
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	// Use WithSandbox for automatic cleanup
	err = client.WithSandbox(ctx, func(ctx context.Context, box *tavor.BoxHandle) error {
		// Run a simple command
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

func manualExample() {
	fmt.Println("\n=== Manual Example: Python Script ===")

	client, err := tavor.NewClient("")
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	// Create box with custom configuration
	config := &tavor.BoxConfig{
		Template: tavor.BoxTemplateBasic,
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
}

func streamingExample() {
	fmt.Println("\n=== Streaming Example: Long Running Command ===")

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

func concurrentExample() {
	fmt.Println("\n=== Concurrent Example: Multiple Commands ===")

	client, err := tavor.NewClient("")
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	box, err := client.CreateBox(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer box.Stop(ctx)

	if err := box.WaitUntilReady(ctx); err != nil {
		log.Fatal(err)
	}

	// Run multiple commands concurrently
	type result struct {
		name   string
		output *tavor.CommandResult
		err    error
	}

	results := make(chan result, 3)

	// Launch concurrent commands
	commands := []struct {
		name string
		cmd  string
	}{
		{"hostname", "hostname"},
		{"date", "date"},
		{"uptime", "uptime"},
	}

	for _, cmd := range commands {
		go func(name, command string) {
			output, err := box.Run(ctx, command, nil)
			results <- result{name: name, output: output, err: err}
		}(cmd.name, cmd.cmd)
	}

	// Collect results
	for i := 0; i < len(commands); i++ {
		res := <-results
		if res.err != nil {
			fmt.Printf("%s failed: %v\n", res.name, res.err)
		} else {
			fmt.Printf("%s: %s", res.name, res.output.Stdout)
		}
	}
}

func errorHandlingExample() {
	fmt.Println("\n=== Error Handling Example ===")

	client, err := tavor.NewClient("")
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	err = client.WithSandbox(ctx, func(ctx context.Context, box *tavor.BoxHandle) error {
		// Command that will fail
		result, err := box.Run(ctx, "exit 42", nil)
		if err != nil {
			return fmt.Errorf("command failed: %w", err)
		}

		if result.ExitCode != 0 {
			fmt.Printf("Command exited with code %d\n", result.ExitCode)
			if result.Stderr != "" {
				fmt.Printf("Error output: %s\n", result.Stderr)
			}
		}

		// Command that will timeout
		opts := &tavor.CommandOptions{
			Timeout: 2000, // 2 seconds
		}

		_, err = box.Run(ctx, "sleep 10", opts)
		if err != nil {
			switch e := err.(type) {
			case *tavor.CommandTimeoutError:
				fmt.Printf("Command timed out after %dms\n", e.Timeout)
			default:
				return fmt.Errorf("unexpected error: %w", err)
			}
		}

		return nil
	}, nil)
	if err != nil {
		log.Printf("Session error: %v", err)
	}
}

func listBoxesExample() {
	fmt.Println("\n=== List Boxes Example ===")

	client, err := tavor.NewClient("")
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	// List all boxes
	boxes, err := client.ListBoxes(ctx)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Found %d boxes:\n", len(boxes))
	for _, box := range boxes {
		fmt.Printf("- Box %s: Status=%s, Created=%s\n",
			box.ID, box.Status, box.InsertedAt.Format(time.RFC3339))
		if len(box.Metadata) > 0 {
			fmt.Printf("  Metadata: %v\n", box.Metadata)
		}
	}
}
