[![Go Reference](https://pkg.go.dev/badge/github.com/on2itsecurity/go-auxo.svg)](https://pkg.go.dev/github.com/on2itsecurity/go-auxo)

# Introduction

This repository contains an AUXO API wrapper in GO.

AUXO API documentation: https://api.on2it.net/v3/doc

## Versions

Check the tags for the most current version.

Version 2 expects a context to be passed as the first argument to each function that makes HTTP calls. This allows users to have more control over the HTTP calls, including timeouts, cancellation, and request tracing. If the context is passed as `nil`, it will use the default (`context.Background()`).

Migrating from version 1.x to 2.x will require adding context as the first parameter to all function calls.

## Using the Auxo API wrapper

### Requirements

- Address of the AUXO portal (API)
- Security token

### Go MOD

When using modules, initiate your (new) project;

```bash
go mod init github.com/projectname
```

Add the API module to `go.mod`, it is recommended to specificly specify the version

```go
module github.com/projectname

go 1.23

require (
   github.com/on2itsecurity/go-auxo v1.0.12
)
```

Download the package.

```bash
go mod vendor
```

## Adding the package to your project

1. Include the library in your projects in the Import.
   ```go
   import (
   	"context"
   	"time"
   	"github.com/on2itsecurity/go-auxo"
   )
   ```

2. Create the APIClient object with Token and Address to connect to.
   ```go
    auxoClient := auxo.NewClient(address, token, debug)
   ```

3. Call the functions, i.e.
   ```go
   // Using default context
   allProtectSurfaces, err := auxoClient.ZeroTrust.GetProtectSurfaces(nil)
   
   // Using context with timeout
   ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
   defer cancel()
   allProtectSurfaces, err := auxoClient.ZeroTrust.GetProtectSurfaces(ctx)
   ```

### Structure

The aim is to support all Auxo API endpoints, currently;

* Asset
* CaseIntegration
* CRM
* Eventflow
* ZeroTrust
* ZTReadiness

These different endpoints can be called with the same `client`, i.e.;

```go
auxoClient.Asset.<action>
auxoClient.CaseIntegration.<action>
auxoClient.CRM.<action>
auxoClient.Eventflow.<action>
auxoClient.ZeroTrust.<action>
auxoClient.ZTReadiness.<action>
```

