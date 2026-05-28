# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**TGDP** - Traffic Generator for Diameter Protocol. 
A Diameter packet generator for network engineers to simulate real-world traffic,
validate system behavior, and troubleshoot Diameter-based applications.

## Build Commands

```bash
make              # Build the project (debug mode with -N -l flags)
make release      # Build release version with static linking
make pkl2go       # Generate Go code from Pkl definitions (Core.pkl)
make mod          # Run go mod tidy and vendor
make clean        # Remove binary and *.pcap files
```

Run the binary: `./tgdp`

## Architecture

```
cmd/tgdp/main.go    # Entry point
internal/           # Internal packages
  cli/              # CLI handling
  config/           # Configuration
  diameter/         # Diameter protocol wrapper
  flags/            # Flag handling
  iface/            # Network interface utilities
  lua/              # Lua scripting integration (avp, l2g, message, peer subpackages)
  repl/             # REPL interactive mode (avp, close, comp, connect, echo, list, receive, script, send, server, verbose subpackages)

pkg/diameter/      # Public Diameter protocol package
  diameter.go       # Core protocol implementation
  avp.go            # AVP definitions
  avpcodec.go       # AVP encoding/decoding
  avpstore.go       # AVP storage
  message.go        # Message handling
  api/              # Interface to Diameter messages as bytes slice
  dict/             # Dictionary (generated from Pkl)
  net/              # Networking layer (node, server, transport)
  diwe/             # Diameter-specific implementations
  pcap/             # PCAP file writing
```

## Key Technologies

- **Cobra** - CLI framework
- **gopacket** - packet processing
- **gopher-lua** - Lua scripting for automation
- **Pkl** - configuration/data definition (generates Go code)
