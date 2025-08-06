# Devento Go SDK Examples

This directory contains examples demonstrating various features of the Devento Go SDK.

## Examples

### Basic Usage (`basic/`)

Demonstrates the simplest way to use Devento with automatic cleanup using `WithSandbox`.

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

Shows how to create and execute Python scripts in a Devento box.

```bash
cd python-script
go run main.go
```

### Web Services (`web/`)

Demonstrates how to expose web services running in a Devento box to the internet with public URLs.

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

1. Set your Devento API key:

   ```bash
   export DEVENTO_API_KEY="sk-devento-your-api-key"
   ```

2. Installed the SDK:

   ```bash
   go get github.com/devento-ai/sdk-go
   ```

## Additional Configuration

You can also configure the SDK using environment variables:

- `DEVENTO_BASE_URL` - Custom API base URL (defaults to <https://api.devento.ai>)
- `DEVENTO_BOX_CPU` - Default CPU cores (e.g., 1, 2)
- `DEVENTO_BOX_MIB_RAM` - Default RAM in MiB (e.g., 128, 256, 512, 1024, 2048)
- `DEVENTO_BOX_TIMEOUT` - Default box timeout in seconds

