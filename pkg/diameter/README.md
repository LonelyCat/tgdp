# Diameter Package

The `diameter` package is the core implementation of the Diameter protocol (RFC 6733) for the TGDP project.
It provides a complete toolset for constructing, parsing, and transporting Diameter messages.

## Quick Start

```go
// Initialize environment
d, _ := diameter.New(diameter.ModeTransaction)
d.LoadDict("dictionary.pkl", dict.FormatPkl)

// Create and send a request
msg, _ := d.NewRequest("S6a", "UL")
peer, _ := d.NewPeer("peer1", "127.0.0.1", 3868, "tcp", 5)
d.SendMessage(peer, msg)
```

## Detailed API Documentation

This package provides a Go implementation of the Diameter protocol (RFC 6733) for the TGDP (Traffic Generator for Diameter Protocol) project.

## Overview

The `diameter` package provides types and functions for creating, encoding, decoding, and sending Diameter protocol messages. It supports TCP and SCTP transports, AVP (Attribute-Value Pair) handling, dictionary-based message construction, and peer-to-peer communication.

## Installation

```go
import "tgdp/pkg/diameter"
```

## Core Concepts

### Diameter Environment

The `Diameter` struct is the central environment that manages:
- Dictionary for AVP/Command/Application definitions
- Peer connections
- AVP storage for message templates
- Codecs for encoding/decoding AVP values

### Key Components

```
pkg/diameter/
├── diameter.go    # Core environment (Diameter struct)
├── message.go     # Diameter message handling
├── avp.go         # AVP (Attribute-Value Pair) handling
├── avpstore.go    # AVP storage and retrieval
├── avpcodec.go    # AVP encoding/decoding
├── dict/          # Dictionary (applications, commands, AVPs)
└── net/           # Networking layer (nodes, transport, server)
```

---

## Core Types

### Diameter

The main environment struct for Diameter protocol operations.

```go
type Diameter struct {
    // Contains: mode, dict, peers, store, codecs, datah, verbLvl, dia2go, rng
}
```

#### Constructor

```go
// New creates a new Diameter environment with the specified mode.
// Mode: diameter.ModeTransaction (0) or diameter.ModeSession (1)
d, err := diameter.New(diameter.ModeTransaction)
```

#### Configuration Methods

| Method | Description |
|--------|-------------|
| `Dict()` | Returns reference to the dictionary |
| `Peers()` | Returns reference to peer nodes |
| `Store()` | Returns reference to AVP storage |
| `Mode()` | Returns current operating mode |
| `SetMode(mode int32)` | Sets operating mode |
| `LoadDict(file string, format int)` | Loads dictionary from file |
| `LoadPeers(file string)` | Loads peer configuration |
| `LoadData(file string)` | Loads AVP data from file |

#### Message Creation

```go
// NewRequest creates a new request message
msg, err := d.NewRequest(appId any, cmdCode any)

// NewAnswer creates a new answer message
msg, err := d.NewAnswer(appId any, cmdCode any)

// NewEmptyMessage creates an empty message
msg, err := d.NewEmptyMessage()
```

#### AVP Methods

```go
// NewAvp creates a new AVP with specified parameters
avp := d.NewAvp(name string, code uint32, flags uint8, vndId uint32, datatype int)

// GetAvp retrieves AVP definition from dictionary
avp, err := d.GetAvp(avpId any) // code (uint32) or name (string)

// CloneAvp creates a deep copy of an AVP
avp := d.CloneAvp(avp *Avp)
```

#### Network Methods

```go
// NewPeer creates a new peer node
peer, err := d.NewPeer(name, addr, port, proto, timeout)

// NewClient creates a new client node with transport
node := d.NewClient(tr transport.ITransport)

// SendCapExchange sends Capabilities-Exchange message
err := d.SendCapExchange(peer *node.Node)

// SendDisconnectPeer sends Disconnect-Peer-Notification
err := d.SendDisconnectPeer(peer *node.Node)

// SendMessage sends a Diameter message to peer
err := d.SendMessage(peer *node.Node, appId, cmd uint32)

// RecvMessage receives a message from peer
msg, err := d.RecvMessage(peer *node.Node, use2 bool)

// ReplyMessage replies to a received message
err := d.ReplyMessage(peer *node.Node, msg *Message)

// BytesToMessage deserializes a byte slice to a Message
msg, err := d.BytesToMessage(data []byte)
```

#### Debug Methods

```go
// VerboseLevel returns current verbosity level
level := d.VerboseLevel()

// SetVerboseLevel sets verbosity (VerboseQuiet, VerboseMsg, VerbosePeer, VerboseCM)
d.SetVerboseLevel(level)

// Verbose prints debug info if level <= current verbosity
d.Verbose(obj diameter.IDebug, level)
```

---

### Message

Represents a Diameter protocol message with header and AVPs.

```go
type Message struct {
    Version  uint8   // Protocol version (1)
    Length   uint32  // Message length in bytes
    AppId    uint32  // Application ID
    Flags    uint8   // Message flags (R, P, E, T)
    CmdCode  uint32  // Command code
    HopByHop uint32  // Hop-by-hop identifier
    EndToEnd uint32  // End-to-end identifier
}
```

#### Message Methods

| Method | Description |
|--------|-------------|
| `AddAvp(avpId any)` | Add AVP to message (by code, name, or *Avp) |
| `RemoveAvp(avpId any)` | Remove AVP from message |
| `GetAvp(avpId any)` | Get first matching AVP |
| `GetAvp2(avpId any, index int)` | Get nth matching AVP (1-based) |
| `GetAvpValue(avpId any)` | Get value of first matching AVP |
| `Reply()` | Create answer message for request |
| `Serialize()` | Convert to wire format bytes |
| `Deserialize(data []byte)` | Parse from wire format |
| `Bytes()` | Get cached wire format |
| `Len()` | Get message length |
| `IsRequest()` | Check if R flag is set |
| `IsProxyable()` | Check if P flag is set |
| `IsError()` | Check if E flag is set |
| `IsRetransmition()` | Check if T flag is set |
| `Print()` | Print human-readable format |
| `Dump(shift ...int)` | Print with indentation |

---

### Avp

Represents a Diameter Attribute-Value Pair.

```go
type Avp struct {
    header dict.Avp  // AVP definition from dictionary
    value  *AvpData  // Encoded value
    env    *Diameter // Environment reference
    length uint32    // Total AVP length
}
```

#### AVP Methods

| Method | Description |
|--------|-------------|
| `Code()` | Get AVP code |
| `Name()` | Get AVP name |
| `Flags()` | Get AVP flags byte |
| `VendorId()` | Get Vendor-ID (0 if not vendor-specific) |
| `Type()` | Get AVP data type |
| `Enum()` | Get enumeration for Enumerated type |
| `Group()` | Get group definition for Grouped type |
| `Data()` | Get raw AvpData |
| `Value()` | Get decoded Go value |
| `Len()` | Calculate total AVP length |
| `SetValue(value *any)` | Set and encode AVP value |
| `Serialize(buf *Buffer)` | Encode to wire format |
| `Deserialize(data []byte)` | Decode from wire format |
| `IsVendorSpec()` | Check if V flag is set |
| `IsMandatory()` | Check if M flag is set |
| `IsProtected()` | Check if P flag is set |
| `IsGrouped()` | Check if type is Grouped |
| `Print()` | Print human-readable format |
| `Dump(shift ...int)` | Print with indentation |

---

### AvpStore

Thread-safe storage for AVP values indexed by AVP code.

```go
type AvpStore struct {
    mu   sync.RWMutex
    data map[uint32][]*Avp
    env  *Diameter
}
```

#### Store Constants

```go
const (
    AvpAppend  = -1  // Append AVPs to existing
    AvpReplace = -2 // Replace AVPs at index
    AvpDelete  = -3 // Delete AVP at index
    AvpPurge   = -4 // Remove all AVPs for code
)
```

#### Store Methods

| Method | Description |
|--------|-------------|
| `Load(id uint32)` | Get all AVPs for code |
| `Store(id uint32, values []*Avp)` | Replace all AVPs for code |
| `Append(id uint32, data []*Avp)` | Add AVPs to existing |
| `Replace(id uint32, index int, data []*Avp)` | Replace at index |
| `Delete(id uint32, index int)` | Delete at index |
| `Purge(id uint32)` | Remove all for code |
| `Range(fn)` | Iterate over all AVPs |
| `LoadFromFile(file string, action, index int)` | Load from YAML |
| `MakeFromYaml(yaml string, action, index int)` | Parse YAML to AVPs |
| `Iter()` | Sequence of AVP slices |
| `Iter2()` | Sequence of (code, AVPs) pairs |

---

## Dictionary

The dictionary package (`pkg/diameter/dict`) provides access to Diameter applications, commands, and AVP definitions.

### Dict

```go
type Dict struct {
    // Thread-safe access to applications, commands, AVPs
}
```

#### Dictionary Methods

```go
// GetApp retrieves application by ID or name
app, err := d.GetApp(appId any) // uint32 or string

// GetCmd retrieves command by code or name within app
cmd, err := d.GetCmd(cmdId any, app *App)

// GetAvp retrieves AVP by code or name
avp, err := d.GetAvp(avpId any)

// GetAppById returns application by numeric ID
app, err := d.GetAppById(appId uint32)

// GetAppByName returns application by name
app, err := d.GetAppByName(appName string)

// GetCmdByCode returns command by numeric code
cmd, err := d.GetCmdByCode(cmdCode uint32, app *App)

// GetCmdByName returns command by name
cmd, err := d.GetCmdByName(cmdName string, app *App)

// GetAvpByCode returns AVP by numeric code
avp, err := d.GetAvpByCode(avpCode uint32)

// GetAvpByName returns AVP by name
avp, err := d.GetAvpByName(avpName string)
```

#### Dictionary Accessors

```go
// AvpFlag returns AVP bit flag definitions (V, M, P)
flags := d.AvpFlag()

// CmdFlag returns Command bit flag definitions (R, P, E, T)
flags := d.CmdFlag()

// AvpDataType returns AVP data type definitions
types := d.AvpDataType()
```

#### Iterators

```go
// Iterate over applications
for app := range d.AppIter() { }

// Iterate with index
for i, app := range d.AppIter2() { }

// Iterate over commands for an app
for cmd := range d.CmdIter(app) { }

// Iterate over all AVPs
for avp := range d.AvpIter() { }
```

---

## Network

### Node

Represents a Diameter peer connection (`pkg/diameter/net/node`).

```go
type Node struct {
    Name       string
    Address    string
    RemotePort int
    LocalPort  int
    Type       string
    Timeout    int
    RouteInfo  RouteInfo
}
```

#### Node Methods

| Method | Description |
|--------|-------------|
| `Connect(sendCe bool)` | Establish connection |
| `Disconnect(sendDp bool)` | Close connection |
| `SendTo(buf []byte)` | Send raw bytes |
| `RecvFrom()` | Receive with interrupt support |
| `RecvFrom2()` | Receive without interrupt |
| `Shutdown()` | Trigger interrupt signal |
| `CollectRouteInfo()` | Query routing table |
| `Transport()` | Get transport type |
| `State()` | Get connection state |
| `SetState(state int32)` | Set connection state |
| `IsOpen()` | Check if connected |
| `IsClient()` | Check if client-side |
| `HasData()` | Check pending data |

#### Node States

```go
const (
    StateClosed       = 0
    StateWaitConnAck  = 1
    StateWaitCEA      = 2
    StateIOpen        = 3  // Initiator open
    StateROpen        = 4  // Responder open
    StateSuspect      = 5
    StateReOpen       = 6
    StateShuttingDown = 7
)
```

#### Transport Types

```go
const (
    TransportUnknown = 0
    TransportSctp    = 1
    TransportTcp     = 2
)
```

### Server

Diameter server with dual-transport support (`pkg/diameter/net/server`).

```go
// Create server
srv := server.New(env interface{}, ctrlChan chan int)

// Start listening
err := srv.Start(addr string, autoRcv bool)

// Stop gracefully
srv.Stop()

// Wait for shutdown
srv.Wait()
```

---

## Constants

### Mode Constants

```go
const (
    ModeTransaction = 0
    ModeSession      = 1
)
```

### Verbose Level Constants

```go
const (
    VerboseQuiet  = 0
    VerboseMsg    = 1
    VerbosePeer   = 2
    VerboseCM     = 3
)
```

### Command Codes (RFC 6733)

```go
const (
    cmdCapExch     = 257 // Capabilities-Exchange
    cmdDevWatchdog = 280 // Device-Watchdog
    cmdDiscPeer    = 282 // Disconnect-Peer-Notification
)
```

### Result Codes

```go
const (
    DiameterSuccess = 2001
)
```

---

## Examples

### Basic Message Creation

```go
// Create Diameter environment
d, err := diameter.New(diameter.ModeTransaction)
if err != nil {
    log.Fatal(err)
}

// Load dictionary
if err := d.LoadDict("Diameter.pkl", dict.FormatPkl); err != nil {
    log.Fatal(err)
}

// Load AVP data
if err := d.LoadData("avps.yaml"); err != nil {
    log.Fatal(err)
}

// Create a request message
msg, err := d.NewRequest("S6a", "UL")
if err != nil {
    log.Fatal(err)
}

// Serialize to wire format
data, err := msg.Serialize()
if err != nil {
    log.Fatal(err)
}

// Print message
msg.Dump()
```

### Working with AVPs

```go
// Create AVP from dictionary
avp, err := d.GetAvp("Session-Id")
if err != nil {
    log.Fatal(err)
}

// Set value
value := "test.session;12345;67890"
if err := avp.SetValue(&value); err != nil {
    log.Fatal(err)
}

// Add to message
if err := msg.AddAvp(avp); err != nil {
    log.Fatal(err)
}

// Get AVP value from message
sessionId, err := msg.GetAvpValue("Session-Id")
if err != nil {
    log.Fatal(err)
}
fmt.Println("Session-ID:", sessionId)
```

### Peer Connection

```go
// Load peer configuration
if err := d.LoadPeers("peers.yaml"); err != nil {
    log.Fatal(err)
}

// Get peer
peer := d.Peers().Get("peer1")
if peer == nil {
    log.Fatal("peer not found")
}

// Connect and perform handshake
if err := peer.CollectRouteInfo(); err != nil {
    log.Fatal(err)
}

if err := peer.Connect(true); err != nil {
    log.Fatal(err)
}

// Send message
if err := d.SendMessage(peer, 16777251, 316); err != nil {
    log.Fatal(err)
}

// Receive response
resp, err := d.RecvMessage(peer, false)
if err != nil {
    log.Fatal(err)
}

resp.Dump()
```

### Using AvpStore

```go
// Load AVPs from YAML
store := d.Store()
if err := store.LoadFromFile("data.yaml", diameter.AvpAppend, 0); err != nil {
    log.Fatal(err)
}

// Iterate over all AVPs
store.Range(func(code uint32, avps []*Avp) bool {
    for _, avp := range avps {
        fmt.Printf("AVP: %s = %v\n", avp.Name(), avp.Value())
    }
    return true
})
```

### Receiving Messages

```go
// Listen for incoming connections
srv := server.New(d, ctrlChan)
if err := srv.Start(":3868", true); err != nil {
    log.Fatal(err)
}

// Handle messages in your OnNetEvent handler
// The RecvHandler processes incoming Diameter messages
```

---

## YAML Format for AVP Data

The `AvpStore` supports loading AVPs from YAML files:

```yaml
# Scalar value
Session-Id: "test.session;12345;67890"
Result-Code: 2001

# Multiple values
Auth-Application-Id:
  - 16777251
  - 4

# Grouped AVPs
Destination-Host:
  host:
    name: example.com
```

---

## Error Types

The package provides specialized errors in `pkg/diameter/diwe`:

- `ErrInvalidMode` - Invalid operating mode
- `ErrUnknownApp` - Unknown application
- `ErrUnknownCmd` - Unknown command
- `ErrUnknownAvp` - Unknown AVP
- `ErrNoValueForReqAvp` - Missing required AVP value
- `ErrInvalidAvpValue` - Invalid AVP value type
- `ErrMissingAvp` - AVP not found in message
- `ErrMsgTooShort` - Message too short
- `ErrConnect` - Connection error
- `ErrSendTo` - Send error
- `ErrRecvFrom` - Receive error

---

## Dependencies

- `tgdp/pkg/diameter/dict` - Dictionary definitions
- `tgdp/pkg/diameter/diwe` - Error types
- `tgdp/pkg/diameter/net/node` - Peer/node management
- `tgdp/pkg/diameter/net/transport` - TCP/SCTP transport
- `tgdp/pkg/diameter/net/server` - Server implementation

---

## See Also

- [RFC 6733](https://tools.ietf.org/html/rfc6733) - Diameter Base Protocol
- [TGDP Project Documentation](../../docs)
