# go-helm-toolkit

HELM client helpers and wrappers used by other Adevinta tools & operators.

## Overview

The `go-helm-toolkit` is a Go library that provides a clean, type-safe interface for interacting with Helm charts and repositories. It abstracts the complexity of Helm CLI operations and provides a consistent API for common Helm operations like installing, upgrading, templating, and managing charts.

## Features

- **Helm 3 Support**: Full support for Helm 3.x operations
- **Type-Safe Interface**: Clean Go interface for Helm operations
- **Chart Discovery**: Utilities to discover and scan chart directories
- **Metadata Loading**: Parse and work with Chart.yaml metadata
- **Test Utilities**: Helper functions for testing Helm chart operations
- **Repository Management**: Add and update Helm repositories
- **Version Management**: Support for multiple Helm versions
- **Flag System**: Flexible flag system for customizing Helm operations

## Installation

```bash
go get github.com/adevinta/go-helm-toolkit
```

## Usage

### Basic Operations

```go
package main

import (
    "io"
    "log"
    
    "github.com/adevinta/go-helm-toolkit"
)

func main() {
    // Get default Helm client
    h, err := helm.Default()
    if err != nil {
        log.Fatal(err)
    }
    
    // Template a chart
    reader, err := h.Template("my-namespace", "my-release", "./my-chart")
    if err != nil {
        log.Fatal(err)
    }
    
    // Install a chart
    err = h.Install("my-namespace", "my-release", "./my-chart", 
        helm.Values("values.yaml"),
        helm.Version("1.0.0"))
    if err != nil {
        log.Fatal(err)
    }
    
    // Upgrade a chart
    err = h.Update("my-namespace", "my-release", "./my-chart")
    if err != nil {
        log.Fatal(err)
    }
    
    // Delete a release
    err = h.Delete("my-namespace", "my-release")
    if err != nil {
        log.Fatal(err)
    }
}
```

### Repository Operations

```go
// Add a repository
err := h.RepoAdd("stable", "https://kubernetes-charts.storage.googleapis.com",
    helm.RepoUsername("user"),
    helm.RepoPassword("pass"))

// Update repositories
err := h.RepoUpdate()
```

### Chart Discovery

```go
// Discover all charts in a directory
for chartDir := range helm.DiscoverChartDirs("/path/to/charts") {
    fmt.Printf("Found chart: %s\n", chartDir)
}

// Discover test files in a chart
for testFile := range helm.DiscoverChartTests("/path/to/chart") {
    fmt.Printf("Found test: %s\n", testFile)
}
```

### Metadata Operations

```go
// Load chart metadata
metadata, err := helm.LoadMetadata("/path/to/chart")
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Chart name: %s\n", metadata.Name)
fmt.Printf("Chart version: %s\n", metadata.Version)
```

### Using Flags

The toolkit provides a flexible flag system for customizing Helm operations:

```go
// Available flags
helm.Debug()                    // Enable debug output
helm.UpgradeInstall()           // Install if not exists
helm.Values("values.yaml")      // Specify values file
helm.RepoUsername("user")       // Repository username
helm.RepoPassword("pass")       // Repository password
helm.Version("1.0.0")          // Specify chart version
```

### Testing

The toolkit includes test utilities for writing Helm chart tests:

```go
package main

import (
    "context"
    "testing"
    
    "github.com/adevinta/go-helm-toolkit"
    "github.com/adevinta/go-helm-toolkit/testutils"
    "github.com/stretchr/testify/require"
)

func TestChartInstallation(t *testing.T) {
    h, err := helm.Default()
    require.NoError(t, err)
    
    // Install and test a chart
    helmtestutils.InstallFilteredHelmChart(t, context.Background(), client, 
        "test-namespace", "release-name", "examples/hello-world", 
        k8stestutils.ExtractObjectName(k8stestutils.WithKind("Deployment"), &name))
}
```

## API Reference

### Core Interface

```go
type Helm interface {
    Template(namespace, release, chart string, flags ...Flag) (io.Reader, error)
    Install(namespace, release, chart string, flags ...Flag) error
    Update(namespace, release, chart string, flags ...Flag) error
    Package(chart string, flags ...Flag) error
    UpdateDeps(chart string) error
    Test(namespace, release string) error
    Delete(namespace, release string) error
    Init() error
    RepoAdd(name, url string, flags ...Flag) error
    RepoUpdate() error
    Version() string
}
```

### Metadata Types

```go
type Metadata struct {
    Name        string                 `json:"name,omitempty"`
    Version     string                 `json:"version,omitempty"`
    Description string                 `json:"description,omitempty"`
    APIVersion  string                 `json:"apiVersion,omitempty"`
    AppVersion  string                 `json:"appVersion,omitempty"`
    // ... other fields
}
```

## Dependencies

This toolkit depends on several Adevinta internal packages:
- `github.com/adevinta/go-k8s-toolkit` - Kubernetes utilities
- `github.com/adevinta/go-system-toolkit` - System utilities
- `github.com/adevinta/go-testutils-toolkit` - Testing utilities

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Submit a pull request

## License

This project is part of Adevinta's internal tooling ecosystem.
