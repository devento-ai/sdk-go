# Tavor Go SDK

> [!WARNING]
> **This package is being renamed to `github.com/devento-ai/sdk-go`**. Please use the new package instead as this one will no longer be maintained. Check out [github.com/devento-ai/sdk-go](https://github.com/devento-ai/sdk-go).
> 
> The current library can still be used with `TAVOR_BASE_URL="https://api.devento.ai"` but will receive no further updates.

The official Go SDK for [Tavor](https://tavor.dev), a cloud sandbox platform that provides secure, isolated execution environments.

## Installation

```bash
go get github.com/tavor-dev/sdk-go
```

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "log"

    tavor "github.com/tavor-dev/sdk-go"
)

func main() {
    // Create client (uses TAVOR_API_KEY env var)
    client, err := tavor.NewClient("")
    if err != nil {
        log.Fatal(err)
    }

    ctx := context.Background()

    // Use WithSandbox for automatic cleanup
    err = client.WithSandbox(ctx, func(ctx context.Context, box *tavor.BoxHandle) error {
        result, err := box.Run(ctx, "echo 'Hello from Tavor!'", nil)
        if err != nil {
            return err
        }

        fmt.Println(result.Stdout)
        return nil
    }, nil)

    if err != nil {
        log.Fatal(err)
    }
}
```

## Authentication

The SDK requires an API key for authentication. You can provide it in two ways:

1. **Environment Variable** (recommended):

   ```bash
   export TAVOR_API_KEY="sk-tavor-xxx"
   ```

2. **Direct Parameter**:

   ```go
   client, err := tavor.NewClient("sk-tavor-xxx")
   ```

## Configuration

### Client Options

```go
httpClient := &http.Client{
    Timeout: 60 * time.Second,
}
client, err := tavor.NewClient("",
    tavor.WithHTTPClient(httpClient),
    tavor.WithBaseURL("https://custom.api.tavor.dev"),
    tavor.WithDebug(true),
)
```

### Logging

The SDK uses Go's standard `log/slog` package for structured logging. You can enable debug logging or provide your own logger:

```go
// Enable built-in debug logging
client, err := tavor.NewClient("", tavor.WithDebug(true))

// Use a custom JSON logger
jsonLogger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
    Level: slog.LevelDebug,
}))
client, err := tavor.NewClient("", tavor.WithLogger(jsonLogger))

// Use a custom logger with additional context
customLogger := slog.New(slog.NewTextHandler(os.Stdout, nil)).With(
    "service", "my-app",
    "version", "1.0.0",
)
client, err := tavor.NewClient("", tavor.WithLogger(customLogger))
```

### Environment Variables

- `TAVOR_API_KEY` - API key for authentication
- `TAVOR_BASE_URL` - Base URL for API (defaults to <https://api.tavor.dev>)
- `TAVOR_BOX_CPU` - Default CPU cores (e.g., 1, 2, integers only)
- `TAVOR_BOX_MIB_RAM` - Default RAM in MiB (e.g., 128, 256, 512, 1024, 2048)
- `TAVOR_BOX_TIMEOUT` - Default box timeout in seconds

## Usage Examples

### Basic Command Execution

```go
ctx := context.Background()

box, err := client.CreateBox(ctx, nil)
if err != nil {
    log.Fatal(err)
}
defer box.Stop(ctx)

if err := box.WaitUntilReady(ctx); err != nil {
    log.Fatal(err)
}

result, err := box.Run(ctx, "pwd", nil)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Working directory: %s\n", result.Stdout)
```

### Using WithSandbox (Automatic Cleanup)

```go
err := client.WithSandbox(ctx, func(ctx context.Context, box *tavor.BoxHandle) error {
    // Box is automatically created and cleaned up
    result, err := box.Run(ctx, "ls -la", nil)
    if err != nil {
        return err
    }

    fmt.Println(result.Stdout)
    return nil
}, nil)
```

### Custom Box Configuration

```go
// Option 1: Use defaults (1 CPU, 1024 MiB RAM)
box, err := client.CreateBox(ctx, nil)
if err != nil {
    log.Fatal(err)
}
defer box.Stop(ctx)

// Option 2: Specify custom resources
config := &tavor.BoxConfig{
    CPU:     2,    // 2 CPU cores
    MibRAM:  2048, // 2 GiB RAM
    Timeout: 3600, // 1 hour
    Metadata: map[string]string{
        "project": "data-analysis",
        "user":    "john.doe",
    },
}

box, err := client.CreateBox(ctx, config)
if err != nil {
    log.Fatal(err)
}
defer box.Stop(ctx)
```

### Streaming Output

```go
opts := &tavor.CommandOptions{
    OnStdout: func(line string) {
        fmt.Printf("[OUT] %s\n", line)
    },
    OnStderr: func(line string) {
        fmt.Printf("[ERR] %s\n", line)
    },
}

result, err := box.Run(ctx, "npm install", opts)
if err != nil {
    log.Fatal(err)
}
```

### Command Timeouts

```go
opts := &tavor.CommandOptions{
    Timeout: 5000, // 5 seconds in milliseconds
}

result, err := box.Run(ctx, "sleep 10", opts)
if err != nil {
    switch e := err.(type) {
    case *tavor.CommandTimeoutError:
        fmt.Printf("Command timed out after %dms\n", e.Timeout)
    default:
        log.Fatal(err)
    }
}
```

### Concurrent Commands

```go
results := make(chan *tavor.CommandResult, 3)
errors := make(chan error, 3)

commands := []string{"hostname", "date", "uptime"}
for _, cmd := range commands {
    go func(command string) {
        result, err := box.Run(ctx, command, nil)
        if err != nil {
            errors <- err
            return
        }
        results <- result
    }(cmd)
}

for i := 0; i < len(commands); i++ {
    select {
    case result := <-results:
        fmt.Printf("%s", result.Stdout)
    case err := <-errors:
        fmt.Printf("Error: %v\n", err)
    }
}
```

### List Existing Boxes

```go
boxes, err := client.ListBoxes(ctx)
if err != nil {
    log.Fatal(err)
}

for _, box := range boxes {
    fmt.Printf("Box %s: %s\n", box.ID, box.Status)
}
```

### Get Existing Box

```go
boxHandle, err := client.GetBox(ctx, "box-123456")
if err != nil {
    log.Fatal(err)
}

result, err := boxHandle.Run(ctx, "echo 'Using existing box'", nil)
if err != nil {
    log.Fatal(err)
}
```

## Error Handling

The SDK provides specific error types for common scenarios:

```go
result, err := box.Run(ctx, "some-command", nil)
if err != nil {
    switch e := err.(type) {
    case *tavor.AuthenticationError:
        log.Fatal("Invalid API key")
    case *tavor.BoxNotFoundError:
        log.Printf("Box %s not found", e.BoxID)
    case *tavor.CommandTimeoutError:
        log.Printf("Command %s timed out", e.CommandID)
    case *tavor.RateLimitError:
        log.Printf("Rate limited, retry after %d seconds", e.RetryAfter)
    case *tavor.APIError:
        log.Printf("API error: %s (status %d)", e.Message, e.StatusCode)
    default:
        log.Printf("Unknown error: %v", err)
    }
}
```

## Resource Configuration

Tavor allows you to configure CPU and RAM resources for your boxes:

### Default Resources

If you don't specify resources, boxes are created with minimal defaults:

- CPU: 1 core
- RAM: 1 GiB (1024 MiB)

### Common Resource Configurations

```go
// Standard (defaults) - for typical workloads
standard := &tavor.BoxConfig{
    CPU:    1,    // 1 CPU core
    MibRAM: 1024, // 1024 MiB RAM
}

// Performance - for demanding tasks
performance := &tavor.BoxConfig{
    CPU:    2,    // 2 CPU cores
    MibRAM: 2048, // 2 GiB RAM
}

// High Memory - for data processing
highMemory := &tavor.BoxConfig{
    CPU:    1,    // 1 CPU core
    MibRAM: 2048, // 2 GiB RAM
}
```

## Web Access

Tavor boxes can expose services to the internet. Each box gets a unique hostname like `uuid.tavor.app`. To access a service running on a specific port inside the VM:

```go
ctx := context.Background()

box, err := client.CreateBox(ctx, nil)
if err != nil {
    log.Fatal(err)
}
defer box.Stop(ctx)

if err := box.WaitUntilReady(ctx); err != nil {
    log.Fatal(err)
}

// Start a web server on port 8080
go func() {
    _, err := box.Run(ctx, "python -m http.server 8080", nil)
    if err != nil {
        log.Printf("Server error: %v", err)
    }
}()

// Give the server a moment to start
time.Sleep(2 * time.Second)

// Get the public URL for port 8080
publicURL, err := box.GetPublicURL(8080)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Web service is accessible at: %s\n", publicURL)
// Output: https://8080-uuid.tavor.app
```

The URL pattern is `https://{port}-{hostname}` where:

- `port` is the port number inside the VM
- `hostname` is the unique hostname assigned to the box

### Port Exposing

You can dynamically expose ports from inside the sandbox to random external ports. This is useful when you need to access services running inside the sandbox but don't know the port in advance or need multiple services:

```go
ctx := context.Background()

box, err := client.CreateBox(ctx, nil)
if err != nil {
    log.Fatal(err)
}
defer box.Stop(ctx)

if err := box.WaitUntilReady(ctx); err != nil {
    log.Fatal(err)
}

// Start a service on port 3000 inside the sandbox
_, err = box.Run(ctx, "python -m http.server 3000 &", nil)
if err != nil {
    log.Fatal(err)
}

// Give the server a moment to start
time.Sleep(2 * time.Second)

// Expose the internal port 3000 to an external port
exposedPort, err := box.ExposePort(ctx, 3000)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Internal port %d is now accessible on external port %d\n",
    exposedPort.TargetPort, exposedPort.ProxyPort)
fmt.Printf("Port mapping expires at: %s\n", exposedPort.ExpiresAt.Format(time.RFC3339))

// You can now access the service using the proxy_port
// For example: http://sandbox-hostname:proxy_port
```

The `ExposePort` method returns an `ExposedPort` struct with:

- `TargetPort` - The port inside the sandbox (what you requested)
- `ProxyPort` - The external port assigned by the system
- `ExpiresAt` - When this port mapping will expire

### Example: Running a Go HTTP Server

```go
err := client.WithSandbox(ctx, func(ctx context.Context, box *tavor.BoxHandle) error {
    // Create a simple Go web server
    serverCode := `
package main
import (
    "fmt"
    "net/http"
)
func main() {
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        fmt.Fprintf(w, "<h1>Hello from Tavor!</h1>")
    })
    fmt.Println("Server starting on port 3000...")
    http.ListenAndServe(":3000", nil)
}
`
    if _, err := box.Run(ctx, fmt.Sprintf("echo '%s' > server.go", serverCode), nil); err != nil {
        return err
    }

    // Start the server in the background
    go box.Run(ctx, "go run server.go", &tavor.CommandOptions{
        OnStdout: func(line string) {
            fmt.Printf("[SERVER] %s\n", line)
        },
    })

    // Wait for server to start
    time.Sleep(3 * time.Second)

    // Get public URL
    url, err := box.GetPublicURL(3000)
    if err != nil {
        return err
    }

    fmt.Printf("Your Go server is accessible at: %s\n", url)

    // Keep the box alive for demonstration
    time.Sleep(30 * time.Second)
    return nil
}, nil)
```

## API Reference

### Client

- `NewClient(apiKey string, opts ...ClientOption) (*Client, error)` - Create a new client
- `CreateBox(ctx context.Context, config *BoxConfig) (*BoxHandle, error)` - Create a new box
- `ListBoxes(ctx context.Context) ([]*Box, error)` - List all boxes
- `GetBox(ctx context.Context, boxID string) (*BoxHandle, error)` - Get existing box
- `WithSandbox(ctx context.Context, fn func(context.Context, *BoxHandle) error, config *BoxConfig) error` - Run function with automatic cleanup

### BoxHandle

- `ID() string` - Get box ID
- `Status() BoxStatus` - Get current status
- `Metadata() map[string]string` - Get metadata
- `Refresh(ctx context.Context) error` - Update status from API
- `WaitUntilReady(ctx context.Context) error` - Wait for box to be running
- `Run(ctx context.Context, command string, opts *CommandOptions) (*CommandResult, error)` - Execute command
- `Stop(ctx context.Context) error` - Terminate the box
- `Close(ctx context.Context) error` - Alias for Stop
- `GetPublicURL(port int) (string, error)` - Get public URL for accessing a service on the specified port
- `ExposePort(ctx context.Context, targetPort int) (*ExposedPort, error)` - Expose a port from inside the sandbox to a random external port

### Types

```go
type BoxConfig struct {
    CPU      int               // Number of CPU cores (e.g., 1, 2)
    MibRAM   int               // RAM in MiB (e.g., 128, 256, 512, 1024)
    Timeout  int               // Timeout in seconds
    Metadata map[string]string // Custom metadata
}

type CommandOptions struct {
    Timeout      int               // Timeout in milliseconds
    PollInterval int               // Poll interval in milliseconds
    OnStdout     func(line string) // Stdout callback
    OnStderr     func(line string) // Stderr callback
}

type CommandResult struct {
    ID       string        // Command ID
    BoxID    string        // Box ID
    Cmd      string        // Command string
    Status   CommandStatus // Final status
    Stdout   string        // Standard output
    Stderr   string        // Standard error
    ExitCode int           // Exit code
}

type ExposedPort struct {
    ProxyPort  int       // External port assigned by the system
    TargetPort int       // Port inside the sandbox
    ExpiresAt  time.Time // When this port mapping expires
}
```

## Best Practices

1. **Always clean up boxes**: Use `defer box.Stop(ctx)` or `WithSandbox`
2. **Set appropriate timeouts**: Default timeout might not suit long-running tasks
3. **Handle errors properly**: Check for specific error types
4. **Use metadata**: Tag boxes for tracking and debugging
5. **Stream output for long operations**: Use callbacks to monitor progress

## License

MIT License - see LICENSE file for details
