package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	devento "github.com/devento-ai/sdk-go"
)

func main() {
	// Create client (uses DEVENTO_API_KEY env var if not provided)
	client, err := devento.NewClient("", devento.WithDebug(true))
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	// Example: Running a simple web server with public access
	err = client.WithSandbox(ctx, func(ctx context.Context, box *devento.BoxHandle) error {
		fmt.Println("Starting web server example...")

		// Create a simple Python web server
		fmt.Println("Creating index.html...")
		htmlContent := `<!DOCTYPE html>
<html>
<head>
    <title>Devento Web Demo</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; }
        h1 { color: #333; }
        .info { background: #f0f0f0; padding: 20px; border-radius: 5px; }
    </style>
</head>
<body>
    <h1>Welcome to Devento!</h1>
    <div class="info">
        <p>This web server is running inside a Devento cloud sandbox.</p>
        <p>The sandbox provides a secure, isolated environment for running code.</p>
        <p>Current time: <span id="time"></span></p>
    </div>
    <script>
        document.getElementById('time').textContent = new Date().toLocaleString();
    </script>
</body>
</html>`

		if _, err := box.Run(ctx, fmt.Sprintf("echo '%s' > index.html", htmlContent), nil); err != nil {
			return fmt.Errorf("failed to create index.html: %w", err)
		}

		// Start Python HTTP server in the background
		fmt.Println("Starting Python HTTP server on port 8000...")
		serverCmd := "python3 -m http.server 8000"

		// Run server in background
		go func() {
			opts := &devento.CommandOptions{
				OnStdout: func(line string) {
					fmt.Printf("[SERVER] %s\n", line)
				},
				OnStderr: func(line string) {
					fmt.Printf("[SERVER ERROR] %s\n", line)
				},
			}
			box.Run(ctx, serverCmd, opts)
		}()

		// Wait for server to start
		time.Sleep(3 * time.Second)

		// Get the public URL
		publicURL, err := box.GetPublicURL(8000)
		if err != nil {
			return fmt.Errorf("failed to get public URL: %w", err)
		}

		fmt.Println("\n" + strings.Repeat("=", 60))
		fmt.Printf("✅ Web server is now publicly accessible!\n")
		fmt.Printf("🌐 URL: %s\n", publicURL)
		fmt.Println(strings.Repeat("=", 60) + "\n")

		// Test the server from inside the box
		fmt.Println("Testing server connectivity...")
		result, err := box.Run(ctx, "curl -s http://localhost:8000 | head -n 5", nil)
		if err != nil {
			fmt.Printf("Warning: Could not test server: %v\n", err)
		} else {
			fmt.Printf("Server response:\n%s\n", result.Stdout)
		}

		// Keep the server running
		fmt.Println("Server will run for 60 seconds. You can visit the URL in your browser.")
		fmt.Println("Press Ctrl+C to stop early.")
		time.Sleep(60 * time.Second)

		return nil
	}, &devento.BoxConfig{
		CPU:     1,
		MibRAM:  512,
		Timeout: 120, // 2 minutes
		Metadata: map[string]string{
			"example": "web-server",
			"type":    "demo",
		},
	})

	if err != nil {
		log.Printf("Error: %v", err)
	}
}
