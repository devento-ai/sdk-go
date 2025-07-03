package main

import (
	"context"
	"fmt"
	"log"
	"time"

	tavor "github.com/tavor-dev/sdk-go"
)

func main() {
	// Create a new Tavor client
	client, err := tavor.NewClient("", tavor.WithDebug(true))
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	// Create a new sandbox
	fmt.Println("Creating a new sandbox...")
	box, err := client.CreateBox(ctx, &tavor.BoxConfig{
		CPU:    1,
		MibRAM: 1024,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer box.Stop(ctx)

	// Wait for the sandbox to be ready
	fmt.Println("Waiting for sandbox to be ready...")
	if err := box.WaitUntilReady(ctx); err != nil {
		log.Fatal(err)
	}

	// Start a simple HTTP server on port 3000
	fmt.Println("Starting a simple HTTP server on port 3000...")
	_, err = box.Run(ctx, `
cat > server.py << 'EOF'
import http.server
import socketserver

PORT = 3000

Handler = http.server.SimpleHTTPRequestHandler

with socketserver.TCPServer(("", PORT), Handler) as httpd:
    print(f"Server running on port {PORT}")
    httpd.serve_forever()
EOF

nohup python3 server.py > /dev/null 2>&1 & disown
	`, nil)
	if err != nil {
		log.Fatal(err)
	}

	// Give the server a moment to start
	time.Sleep(2 * time.Second)

	// Expose port 3000
	fmt.Println("Exposing port 3000...")
	exposedPort, err := box.ExposePort(ctx, 3000)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Port exposed successfully!\n")
	fmt.Printf("  Target port: %d\n", exposedPort.TargetPort)
	fmt.Printf("  Proxy port: %d\n", exposedPort.ProxyPort)
	fmt.Printf("  Expires at: %s\n", exposedPort.ExpiresAt.Format(time.RFC3339))

	// You can now access your service from outside the sandbox
	// using the proxy_port on the sandbox's hostname

	fmt.Println("\nKeeping sandbox alive for 30 seconds...")
	fmt.Println("You can test the exposed port during this time.")
	time.Sleep(30 * time.Second)

	fmt.Println("Done!")
}
