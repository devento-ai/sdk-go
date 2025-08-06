package main

import (
	"context"
	"fmt"
	"log"
	"time"

	devento "github.com/devento-ai/sdk-go"
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

	// Example 6: Resource allocation strategies
	resourceAllocationExample()
}

func basicExample() {
	fmt.Println("=== Basic Example: Hello World ===")

	// Create client (uses DEVENTO_API_KEY env var if not provided)
	client, err := devento.NewClient("", devento.WithDebug(true))
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	// Use WithSandbox for automatic cleanup
	// Using nil config (defaults: 1 CPU, 1024 MiB RAM)
	err = client.WithSandbox(ctx, func(ctx context.Context, box *devento.BoxHandle) error {
		// Run a simple command
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

func manualExample() {
	fmt.Println("\n=== Manual Example: Python Script ===")

	client, err := devento.NewClient("")
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	// Create box with custom configuration
	// Python workloads benefit from more memory
	config := &devento.BoxConfig{
		CPU:    1,
		MibRAM: 2048,
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

	client, err := devento.NewClient("")
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	// For streaming operations, we might want slightly less memory
	config := &devento.BoxConfig{
		CPU:    1,   // Default CPU is fine for simple streaming
		MibRAM: 512, // 512 MiB RAM to handle buffering
	}

	err = client.WithSandbox(ctx, func(ctx context.Context, box *devento.BoxHandle) error {
		// Run a command with streaming output
		opts := &devento.CommandOptions{
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
	}, config)
	if err != nil {
		log.Printf("Error: %v", err)
	}
}

func concurrentExample() {
	fmt.Println("\n=== Concurrent Example: Multiple Commands ===")

	client, err := devento.NewClient("")
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	// For concurrent operations, allocate more resources
	config := &devento.BoxConfig{
		CPU:    1,   // 2 full CPU cores for parallel execution
		MibRAM: 512, // 512 MiB RAM
		Metadata: map[string]string{
			"purpose": "concurrent-demo",
		},
	}

	box, err := client.CreateBox(ctx, config)
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
		output *devento.CommandResult
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

	client, err := devento.NewClient("")
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	// Error handling doesn't need extra resources - use defaults
	err = client.WithSandbox(ctx, func(ctx context.Context, box *devento.BoxHandle) error {
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
		opts := &devento.CommandOptions{
			Timeout: 2000, // 2 seconds
		}

		_, err = box.Run(ctx, "sleep 10", opts)
		if err != nil {
			switch e := err.(type) {
			case *devento.CommandTimeoutError:
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

	client, err := devento.NewClient("")
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

func resourceAllocationExample() {
	fmt.Println("\n=== Resource Allocation Strategies ===")

	client, err := devento.NewClient("")
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	// Strategy 1: Minimal resources for simple tasks
	fmt.Println("\n1. Minimal Configuration (defaults)")
	err = client.WithSandbox(ctx, func(ctx context.Context, box *devento.BoxHandle) error {
		// Default: 1 CPU, 1024 MiB RAM
		result, err := box.Run(ctx, "echo 'Minimal resources work great for simple tasks!'", nil)
		if err != nil {
			return err
		}
		fmt.Printf("   Output: %s", result.Stdout)
		return nil
	}, nil)
	if err != nil {
		log.Printf("Error: %v", err)
	}

	// Strategy 2: Moderate resources for typical workloads
	fmt.Println("\n2. Moderate Configuration")
	moderateConfig := &devento.BoxConfig{
		CPU:    1,
		MibRAM: 512,
		Metadata: map[string]string{
			"tier": "moderate",
		},
	}
	err = client.WithSandbox(ctx, func(ctx context.Context, box *devento.BoxHandle) error {
		// Good for most scripting tasks
		result, err := box.Run(ctx, "free -m | grep Mem", nil)
		if err != nil {
			return err
		}
		fmt.Printf("   Memory info: %s", result.Stdout)
		return nil
	}, moderateConfig)
	if err != nil {
		log.Printf("Error: %v", err)
	}

	// Strategy 3: High resources for intensive workloads
	fmt.Println("\n3. High Performance Configuration")
	highPerfConfig := &devento.BoxConfig{
		CPU:    2,    // 2 CPU cores
		MibRAM: 2048, // 2 GiB RAM
		Metadata: map[string]string{
			"tier":     "performance",
			"use-case": "data-processing",
		},
	}

	box, err := client.CreateBox(ctx, highPerfConfig)
	if err != nil {
		log.Fatal(err)
	}
	defer box.Stop(ctx)

	if err := box.WaitUntilReady(ctx); err != nil {
		log.Fatal(err)
	}

	// Show system resources
	result, err := box.Run(ctx, "nproc", nil)
	if err == nil {
		fmt.Printf("   CPU cores available: %s", result.Stdout)
	}

	result, err = box.Run(ctx, "free -m | grep Mem | awk '{print $2}'", nil)
	if err == nil {
		fmt.Printf("   Total memory (MiB): %s", result.Stdout)
	}

	// Strategy 4: Dynamic resource allocation based on workload
	fmt.Println("\n4. Dynamic Resource Selection")
	workloadTypes := []struct {
		name   string
		config *devento.BoxConfig
		desc   string
	}{
		{
			name: "web-scraping",
			config: &devento.BoxConfig{
				CPU:    1,   // Minimal CPU
				MibRAM: 256, // Moderate RAM for DOM parsing
			},
			desc: "Light CPU, moderate RAM",
		},
		{
			name: "ml-inference",
			config: &devento.BoxConfig{
				CPU:    2,    // Multiple cores for parallel processing
				MibRAM: 4096, // 4 GiB RAM for models
			},
			desc: "High CPU and RAM",
		},
		{
			name: "build-tasks",
			config: &devento.BoxConfig{
				CPU:    1,    // Single core is often sufficient
				MibRAM: 1024, // 1 GiB RAM for compilation
			},
			desc: "Balanced resources",
		},
	}

	for _, wt := range workloadTypes {
		fmt.Printf("\n   Workload '%s' (%s):\n", wt.name, wt.desc)
		fmt.Printf("   - CPU: %d cores\n", wt.config.CPU)
		fmt.Printf("   - RAM: %d MiB\n", wt.config.MibRAM)
	}
}
