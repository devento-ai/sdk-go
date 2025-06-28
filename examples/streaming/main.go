package main

import (
	"context"
	"fmt"
	"log"
	"os"

	tavor "github.com/tavor-dev/sdk-go"
)

func main() {
	fmt.Println("Testing SSE streaming implementation...\n")

	client, err := tavor.NewClient(os.Getenv("TAVOR_API_KEY"))
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	err = client.WithSandbox(ctx, func(ctx context.Context, box *tavor.BoxHandle) error {
		fmt.Printf("Box %s is ready!\n\n", box.ID())

		fmt.Println("Test 1: Basic streaming output")
		fmt.Println("==============================")

		opts := &tavor.CommandOptions{
			OnStdout: func(line string) {
				fmt.Printf("[STDOUT] %s\n", line)
			},
			OnStderr: func(line string) {
				fmt.Printf("[STDERR] %s\n", line)
			},
		}

		result, err := box.Run(ctx, `for i in {1..3}; do echo "Line $i"; sleep 0.5; done`, opts)
		if err != nil {
			return err
		}

		fmt.Printf("\nTest 1 completed! Exit code: %d\n\n", result.ExitCode)

		fmt.Println("Test 2: Mixed stdout and stderr")
		fmt.Println("================================")

		opts2 := &tavor.CommandOptions{
			OnStdout: func(line string) {
				fmt.Printf("[OUT] %s\n", line)
			},
			OnStderr: func(line string) {
				fmt.Printf("[ERR] %s\n", line)
			},
		}

		result, err = box.Run(ctx, `echo "This is stdout"; >&2 echo "This is stderr"; echo "More stdout"`, opts2)
		if err != nil {
			return err
		}

		fmt.Printf("\nTest 2 completed! Exit code: %d\n\n", result.ExitCode)

		fmt.Println("Test 3: Non-streaming execution")
		fmt.Println("================================")

		result, err = box.Run(ctx, `echo "Hello from non-streaming"`, nil)
		if err != nil {
			return err
		}

		fmt.Printf("Result stdout: %s", result.Stdout)
		fmt.Printf("Result stderr: %s", result.Stderr)
		fmt.Printf("Exit code: %d\n", result.ExitCode)

		fmt.Println("\nAll tests completed successfully!")
		return nil
	}, &tavor.BoxConfig{
		CPU:    1,
		MibRAM: 1024,
	})

	if err != nil {
		log.Fatalf("Test failed: %v", err)
	}
}
