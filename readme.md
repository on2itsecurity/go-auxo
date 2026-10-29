[![Go Reference](https://pkg.go.dev/badge/github.com/on2itsecurity/go-auxo.svg)](https://pkg.go.dev/github.com/on2itsecurity/go-auxo)

# Introduction

This repository contains an AUXO API wrapper in GO.

AUXO API documentation: https://api.on2it.net/v3/doc

## Versions

Check the tags for the most current version.

## Version 2.2.x

Version 2.2 introduces `ZeroTrust.GetHistory`, which returns the current content of all protect surfaces, states and locations with their recorded versions since a given moment, plus helpers to work with it:

* `History.Changes()` lists every recorded change (create, update or delete) with the changed fields, oldest first.
* `History.AsOf(t)` reconstructs the content of all entities at moment `t`.
* `ToProtectSurface`, `ToState` and `ToLocation` convert a version back to an object that can be written, i.e. to roll back.

Measure assignment, implementation and evidence now also support a `Comment`.

The history is keyed by uniqueness key, not by ID. Protect surfaces carry a `CreateDate`, `AsOf` leaves out protect surfaces created after `t` together with their states. States and locations have no creation date, so those created later are still included. A protect surface with an unknown creation time has `CreateDate` `zerotrust.CreationRecordedFrom` (16 September 2026). The content before the oldest recorded version is unknown (`Exact` is `false` in `AsOf`).

```go
ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
defer cancel()
history, err := auxoClient.ZeroTrust.GetHistory(ctx, time.Now().AddDate(0, -1, 0))
pointInTime, err := history.AsOf(time.Now().AddDate(0, 0, -7))
```

## Version 2.1.x (Breaking Change)

Version 2.1 introduces GetCases, but also changes the primary contact email to a pointer, which might be null from this version onwards.

### Version 2.x (Breaking Changes)

Version 2 introduces **breaking changes** that require code modifications when upgrading from v1.x:

**Key Changes:**
- **Context Support**: All functions that make HTTP calls now require a `context.Context` as the first parameter
- **Timeout Control**: Timeout is now controlled via context instead of client timeout
- **Better Cancellation**: HTTP requests can be cancelled using context cancellation

**Migration from v1.x to v2.x:**
1. Add `context` as the first parameter to all API calls
2. Update your imports to include `"context"`
3. Use `nil` for default behavior or pass your own context for custom timeout/cancellation

**Before (v1.x):**
```go
protectSurfaces, err := auxoClient.ZeroTrust.GetProtectSurfaces()
```

**After (v2.x):**
```go
// Using nil (default behavior)
protectSurfaces, err := auxoClient.ZeroTrust.GetProtectSurfaces(nil)

// Using custom context with timeout
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()
protectSurfaces, err := auxoClient.ZeroTrust.GetProtectSurfaces(ctx)
```

### Version 1.x (Legacy)

Version 1.x is the legacy version without context support. Use v1.x tags if you need the old API without breaking changes.

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

go 1.24

require (
   github.com/on2itsecurity/go-auxo/v2 v2.0.0
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
   	"github.com/on2itsecurity/go-auxo/v2"
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

