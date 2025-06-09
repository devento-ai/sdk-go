# Tavor Go SDK Examples

This directory contains examples demonstrating various features of the Tavor Go SDK.

## Examples

### Basic Usage (`basic/`)

Demonstrates the simplest way to use Tavor with automatic cleanup using `WithSandbox`.

```bash
cd basic
go run main.go
```

### Streaming Output (`streaming/`)

Shows how to stream command output in real-time using callbacks.

```bash
cd streaming
go run main.go
```

### Concurrent Commands (`concurrent/`)

Demonstrates running multiple commands in parallel within a single box.

```bash
cd concurrent
go run main.go
```

### Python Script Execution (`python-script/`)

Shows how to create and execute Python scripts in a Tavor box.

```bash
cd python-script
go run main.go
```

### Web Services (`web/`)

Demonstrates how to expose web services running in a Tavor box to the internet with public URLs.

```bash
cd web
go run main.go
```

### Full Example (`main.go`)

A comprehensive example that demonstrates all SDK features including:

- Basic usage with automatic cleanup
- Manual box management
- Streaming output
- Concurrent operations
- Error handling
- Listing boxes

```bash
go run main.go
```

## Prerequisites

Before running any example, make sure you have:

1. Set your Tavor API key:

   ```bash
   export TAVOR_API_KEY="sk-tavor-your-api-key"
   ```

2. Installed the SDK:

   ```bash
   go get github.com/tavor-dev/sdk-go
   ```

## Additional Configuration

You can also configure the SDK using environment variables:

- `TAVOR_BASE_URL` - Custom API base URL (defaults to <https://api.tavor.dev>)
- `TAVOR_BOX_CPU` - Default CPU cores (e.g., 1, 2)
- `TAVOR_BOX_MIB_RAM` - Default RAM in MiB (e.g., 128, 256, 512, 1024, 2048)
- `TAVOR_BOX_TIMEOUT` - Default box timeout in seconds

