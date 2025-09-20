# sslkeylogfile

A Go package that provides SSL/TLS key logging functionality for debugging purposes, compatible with Wireshark and other tools that support the `SSLKEYLOGFILE` environment variable.

## Features

- **Zero-overhead when disabled**: All functions are no-ops unless `SSLKEYLOGFILE` is set
- **Thread-safe logging**: Supports multiple concurrent writers with per-config file management
- **Multiple integration options**: Create custom configs, transports, or patch the default transport
- **Automatic build integration**: Use `sslkeylog-go.sh` to inject SSL key logging into any Go main package

## Usage

### Environment Setup

Set the `SSLKEYLOGFILE` environment variable to specify where SSL/TLS keys should be logged:

```bash
export SSLKEYLOGFILE=/path/to/your/sslkeylogfile.log
```

### Manual Integration

```go
import (
    "crypto/tls"
    "net/http"

    "github.com/fiam/sslkeylogfile"
)

// Option 1: Custom TLS config
tlsConfig, err := sslkeylogfile.NewTLSConfig()
if err != nil {
    panic(err)
}

// Option 2: Custom HTTP transport
transport, err := sslkeylogfile.NewTransport()
if err != nil {
    panic(err)
}
client := &http.Client{Transport: transport}

// Option 3: Patch default transport
if err := sslkeylogfile.PatchDefaultTransport(); err != nil {
    panic(err)
}
```

### Automatic Integration with sslkeylog-go.sh

The `sslkeylog-go.sh` script provides seamless integration by automatically injecting the autopatch import into your main packages during build:

```bash
# Build with SSL key logging enabled
./sslkeylog-go.sh build ./cmd/myapp

# Install with SSL key logging enabled
./sslkeylog-go.sh install ./...

# Works with any go build/install arguments
./sslkeylog-go.sh build -ldflags="-s -w" ./cmd/myapp
```

The script:
- Automatically detects `package main` directories in your build arguments
- Injects `import _ "github.com/fiam/sslkeylogfile/autopatch"` using Go's overlay feature
- Requires no code changes to your application
- Fails safely if no main packages are found

#### Script Usage

```bash
./sslkeylog-go.sh {build|install} [go-build-args...]
```

The script accepts the same arguments as `go build` or `go install`, but automatically adds SSL key logging to any main packages being built.

## How It Works

- **File Management**: Creates separate log files per `tls.Config` to support concurrent usage
- **File Naming**: Uses the `SSLKEYLOGFILE` pattern with sequence numbers (e.g., `sslkeylogfile`, `sslkeylogfile.1`, etc.)
- **Format Compatibility**: Outputs keys in the standard format expected by Wireshark and similar tools
- **Autopatch**: The `autopatch` package automatically calls `PatchDefaultTransport()` on import

## Security Considerations

⚠️ **Warning**: This package logs SSL/TLS keys which can be used to decrypt network traffic. Only use in development/debugging environments. Never use in production with sensitive data.

## License

See LICENSE file for details.