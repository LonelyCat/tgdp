# CLAUDE.md - Diameter Package

This file provides guidance for working with the `pkg/diameter` package.

## Purpose
The `diameter` package implements the core Diameter protocol (RFC 6733).
It handles the lifecycle of Diameter messages, AVP encoding/decoding (via codecs), dictionary management, and network peer communication.

## Key Architecture
- **`Diameter` (Environment)**: The central coordinator. Manages the dictionary, peer list, AVP store, and codecs.
- **`Message`**: Represents a Diameter packet. Handles serialization to/from wire format.
- **`Avp`**: Represents a single Attribute-Value Pair. Relies on codecs for value conversion.
- **`AvpStore`**: A thread-safe repository for AVP values, used for templating messages.
- **`Dict` (`pkg/diameter/dict`)**: Provides metadata about Applications, Commands, and AVPs.
- **`Net` (`pkg/diameter/net`)**: Handles the transport layer (TCP/SCTP) and peer (`Node`) management.

## Critical Constraints & Patterns
- **Alignment**: Diameter AVPs must be padded to 4-byte boundaries (see `alignTo4` in `avp.go`).
- **Codecs**: AVP value conversions are handled by `CodecFuncs`. New AVP types require registering a codec in `LoadDict`.
- **Wire Format**: Big-endian is used for all numeric fields.
- **Memory**: Message serialization is cached in `Message.bytes` to avoid redundant work.

## Common Tasks
- **Adding a new Application**: Update the Pkl definitions in `pkg/diameter/dict/pkl/apps.pkl`.
- **Adding a new AVP**: Update the Pkl definitions in `pkg/diameter/dict/pkl/avps.pkl`.
- **Creating a Message**: Use `d.NewRequest(...)` or `d.NewAnswer(...)`.
- **Setting AVP Values**: Use `avp.SetValue(value)`, which validates the Go type against the dictionary definition.

## Testing
- Run tests: `go test ./pkg/diameter/...`

## Current limitations
- SCTP not yet supported on MacOS.
