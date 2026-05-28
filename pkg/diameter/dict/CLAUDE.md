# CLAUDE.md - pkg/diameter/dict

This file provides guidance for working with the Diameter dictionary package.

## Overview

The `dict` package provides dictionary handling for the Diameter protocol. It manages:
- **Applications** (Auth/Acc applications)
- **Commands** (Diameter messages like CER, CEA, DPR, DPA, etc.)
- **AVPs** (Attribute-Value Pairs)

## Key Files

| File | Description |
|------|-------------|
| `dict.go` | Main dictionary implementation with lookup methods |
| `dict_test.go` | Tests for dictionary functionality |
| `init.pkl.go` | Generated Pkl initialization code |
| `Core.pkl.go`, `App.pkl.go`, `Avp.pkl.go`, etc. | Generated type definitions from Pkl |
| `pkl/` | Pkl source files for dictionary definitions |

## Usage

```go
import "tgdp/pkg/diameter/dict"

// Load dictionary from Pkl file
d := dict.New(core)
err := d.LoadFromFile("Diameter.pkl", dict.FormatPkl)

// Query dictionary
app, _ := d.GetApp("S6a")
cmd, _ := d.GetCmd("UL", app)
avp, _ := d.GetAvp("Session-Id")
```

## Key Methods

- `GetApp(appId)` - Get application by ID (uint32) or name (string)
- `GetCmd(cmdId, app)` - Get command by code or name for specific app
- `GetAvp(avpId)` - Get AVP by code or name
- `AppIter()`, `CmdIter(app)`, `AvpIter()` - Iterate over dictionary items

## Pkl Integration

Dictionary data is defined in Pkl and generated to Go:
- `pkl/dictionary.pkl` - Main dictionary source
- Run `make pkl2go` to regenerate Go code from Pkl
