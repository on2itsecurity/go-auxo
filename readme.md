[![Go Reference](https://pkg.go.dev/badge/github.com/on2itsecurity/go-auxo.svg)](https://pkg.go.dev/github.com/on2itsecurity/go-auxo)

# Introduction

This repository contains an AUXO API wrapper in GO.

AUXO API documentation: https://api.on2it.net/v3/doc

## Versions
Only "major" versions are listed here, check the releases/tags for the most current version.

- Version 1.0.0 - First release

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

go 1.19

require (
   github.com/on2itsecurity/go-auxo v1.0.0
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
   	"github.com/on2itsecurity/go-auxo"
   )
   ```

2. Create the APIClient object with Token and Address to connect to.
   ```go
    auxoClient := auxo.NewClient(address, token, debug)
   ```

3. Call the functions, i.e.
   ```go
   allProtectSurfaces, err := auxoClient.ZeroTrust.GetProtectSurfaces()
   ```

### Structure

The aim is to support all Auxo API endpoints, currently;

* Asset
* CRM
* Eventflow
* ZeroTrust
* ZTReadiness

These different endpoints can be called with the same `client`, i.e.;

```go
auxoClient.Asset.<action>
auxoClient.CRM.<action>
auxoClient.Eventflow.<action>
auxoClient.ZeroTrust.<action>
auxoClient.ZTReadiness.<action>
```

