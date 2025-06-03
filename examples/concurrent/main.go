package main

import (
	"context"
	"fmt"
	"log"

	tavor "github.com/tavor-dev/sdk-go"
)

func main() {
	client, err := tavor.NewClient("", tavor.WithDebug(true))
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	// Example 1: Minimal usage - relies on defaults (1 CPU, 1024 MiB RAM)
	// box, err := client.CreateBox(ctx, nil)

	// Example 2: Explicit resource specification for concurrent workloads
	// Using more resources since we're running multiple commands concurrently
	config := &tavor.BoxConfig{
		CPU:    2,   // 2 CPU cores for better concurrent performance
		MibRAM: 512, // 512 MiB RAM
		Metadata: map[string]string{
			"purpose": "concurrent-commands",
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
		{"id", "id"},
		{"cat /etc/passwd", "cat /etc/passwd"},
	}

	for _, cmd := range commands {
		go func(name, command string) {
			output, err := box.Run(ctx, command, nil)
			results <- result{name: name, output: output, err: err}
		}(cmd.name, cmd.cmd)
	}

	for range commands {
		res := <-results
		if res.err != nil {
			fmt.Printf("%s failed: %v\n", res.name, res.err)
		} else {
			fmt.Printf("%s: %s; %s", res.name, res.output.Stdout, res.output.Stderr)
		}
	}
}
