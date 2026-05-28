# CLAUDE.md - net package

This file provides guidance for working with the Diameter network package (`pkg/diameter/net`).

## Package Overview

The `net` package provides the Diameter protocol networking layer with support for TCP and SCTP transports, peer/node management, and server functionality for the TGDP project.

## Directory Structure

```
net/
├── node/           # Peer/node management
│   ├── node.go     # Node struct and peer operations
│   └── nodes.go    # Nodes collection manager
├── server/         # Diameter server implementation
│   └── server.go   # Server with dual-transport support
└── transport/      # Transport abstraction layer
    ├── transport.go   # ITransport and IListener interfaces
    ├── tcp.go         # TCP transport
    ├── sctp.go        # SCTP transport (common code)
    ├── sctp_linux.go  # Linux SCTP implementation
    ├── sctp_darwin.go # macOS SCTP (not implemented)
    └── error.go       # Transport error types
```

## Key Types

### Node (`node` package)

Represents a Diameter peer connection with full lifecycle management:

**Core Fields:**
- `Name` - Peer identifier
- `Address` - Remote IP address or hostname
- `RemotePort` - Remote TCP/SCTP port (default 3868)
- `LocalPort` - Local TCP/SCTP port
- `Type` - Peer type (e.g., "MME", "HSS")
- `Timeout` - Connection timeout in seconds
- `RouteInfo` - Network routing details (interface, gateway, IPs)

**Transport & State:**
- `tr` - Transport layer (`transport.ITransport`)
- `state` - Atomic state machine
- `client` - Client vs server flag
- `parent` - Reference to parent `Nodes` collection

**Methods:**
- `Connect()` - Establish connection, send Capabilities-Exchange Request
- `Disconnect()` - Send Disconnect-Peer-Request, close connection
- `Close()` - Close transport, cleanup channels
- `SendTo(data []byte)` - Send raw bytes to peer
- `RecvFrom(wait bool)` - Receive with interrupt signal support
- `IsOpen()` - Check if connection is open (IOpen or ROpen)
- `IsClient()` - Returns true for client connections
- `IsClosed()` - Returns true if closed
- `SetTimeout(timeout int)` - Apply timeout to transport
- `Transport()` - Get transport interface
- `Interrupt()` - Send interrupt to unblock RecvFrom
- `Context()` - Get node's context for cancellation
- `GetRouteInfo()` - Query routing table for local interface/gateway
- `Lock/Unlock/TryLock` - Mutex operations

### Nodes (`node` package)

Thread-safe collection for managing multiple peers:

**Methods:**
- `NewNodes()` - Create empty collection
- `NewPeer(name, addr, port, proto, timeout, diaApi)` - Create and add server-side peer
- `NewPeerEx(tr, diaApi, ucb)` - Create and add client from accepted transport
- `GetByName(name)` - Get peer by name (case-insensitive)
- `Remove(name)` - Remove peer by name
- `DisconnectAll(clients bool)` - Disconnect all or client-only peers
- `InterruptAll()` - Send interrupt to all peers
- `Iter()` / `Iter2()` - Iterator over peers
- `LoadFromFile(yamlFile, diaApi)` - Load peers from YAML config
- `Print()` / `Dump()` - Print all peers to stdout

### Server (`server` package)

Dual-transport Diameter server with worker pool:

**Key Features:**
- Simultaneous SCTP and TCP listeners
- Semaphore-based worker pool (max 103 workers)
- Atomic state management
- Structured logging via slog
- Auto-reply capability

**Methods:**
- `New(env *diameter.Diameter) *Server` - Create server
- `Start(addr string, autoReply bool) error` - Start listening
- `Shutdown()` - Graceful shutdown
- `Wait()` - Wait for completion
- `State() / SetState()` - State operations
- `IsRunning() / IsStopped()` - State checks
- `VerboseLevel() / SetVerboseLevel()` - Verbosity control
- `Verbose(level, msg, attrs...)` - Log with level check
- `Dump()` - Print status to stdout
- `Autoreply() / SetAutoreply()` - Auto-reply mode

### ITransport (`transport` package)

Abstract interface for network transports (SCTP or TCP):

```go
type ITransport interface {
    Connect(netip.Addr, int, netip.Addr, int) error
    Close() error
    SetTimeout(int) error
    Send(buf []byte) error
    Recv() ([]byte, error)
    IsConnected() bool
    RemoteAddr() string
    LocalAddr() string
    RemoteIp() netip.Addr
    LocalIp() netip.Addr
    RemotePort() int
    LocalPort() int
    Error() error
    Name() string
    Type() int
}
```

### IListener (`transport` package)

Interface for transport listeners:

```go
type IListener interface {
    Create(string) error
    Accept() (ITransport, error)
    Close() error
    Ready() bool
    Name() string
    Uri() string
}
```

## Constants

### Node State (node package)
```go
StateClosed       = 0  // Closed connection
StateWaitConnAck  = 1  // Waiting for connection ack
StateWaitCEA      = 2  // Waiting for CEA
StateIOpen        = 3  // Open as initiator
StateROpen        = 4  // Open as responder
StateSuspect      = 5  // Peer suspect
StateReOpen       = 6  // Reconnecting
StateShuttingDown = 7  // Shutdown in progress
StateWaitData     = 8  // Waiting for data
```

### Transport Type (node package)
```go
TransportUnknown = 0
TransportSctp    = 1
TransportTcp     = 2
```

### Configuration Keys (node package)
```go
ConfKeyName      = "name"
ConfKeyAddress   = "address"
ConfKeyPort      = "port"
ConfKeyTransport = "transport"
ConfKeyTimeout   = "timeout"
```

### Server State (server package)
```go
StateStopped  = 0  // Server not running
StateRunning  = 1  // Accepting connections
StateShutdown = 2  // Shutting down
```

### Verbosity (server package)
```go
Quiet = 0  // No output
Error = 1  // Errors only
Warn  = 2  // Warnings and above
Info  = 3  // Informational and above
Debug = 4  // All messages
```

### Transport (transport package)
```go
DefaultTimeout   = 10    // seconds
DefaultPort      = 3868  // Diameter default
DefaultProtocol  = "sctp"
```

## Usage Examples

### Creating a Node
```go
diam := diameter.New(env)
nodes := node.NewNodes()

peer, err := nodes.NewPeer(
    "peer1",           // name
    "192.168.1.1",     // address
    3868,              // port
    "sctp",            // transport
    30,                // timeout
    diam,              // diameter API
)
if err != nil {
    // handle error
}
err = peer.Connect()
```

### Creating Transport Directly
```go
// TCP
tcpTr, err := transport.New("tcp")
if err != nil {
    // handle error
}
err = tcpTr.Connect(remoteIP, 3868, localIP, 0)

// SCTP (Linux only)
sctpTr, err := transport.New("sctp")
if err != nil {
    // handle error
}
err = sctpTr.Connect(remoteIP, 3868, localIP, 0)
```

### Server Setup
```go
diam := diameter.New(env)
srv := server.New(diam)

// Start listening
err := srv.Start(":3868", true)  // autoReply = true
if err != nil {
    // handle error
}

// Wait for shutdown
srv.Wait()
```

### YAML Configuration
```yaml
peer1:
  address: 192.168.1.1
  port: 3868
  transport: sctp
  timeout: 30

peer2:
  address: 192.168.1.2
  transport: tcp
  timeout: 20
```

## State Machine

**Node States:**
- `Closed` → `WaitConnAck` → `WaitCEA` → `IOpen` (client) or `ROpen` (server)
- Any open state can transition to `ShuttingDown` → `Closed`
- During receive: any open state → `WaitData` → previous state

**Server States:**
- `Stopped` → `Running` (after Start)
- `Running` → `Shutdown` (on Shutdown)
- `Shutdown` → `Stopped` (after Wait)

## Known Limitations

- **macOS**: SCTP transport is not supported (use TCP)
- **SendTo()**: Accepts `[]byte`, not `*diameter.Message` - caller must serialize
- **Max Clients**: Server limits to 100 concurrent clients via worker pool
- **Buffer Size**: 64KB receive buffer (sufficient for Diameter messages)

## Dependencies

- `tgdp/pkg/diameter` - Diameter environment and API
- `tgdp/pkg/diameter/net/transport` - Transport abstraction
- `tgdp/pkg/diameter/diwe` - Custom error types
- `github.com/libp2p/go-netroute` - Network route lookup
- `github.com/ishidawataru/sctp` - SCTP support (Linux)

## Related Files

- `./node/CLAUDE.md` - Detailed node package documentation
- `./server/CLAUDE.md` - Detailed server package documentation
- `./transport/CLAUDE.md` - Detailed transport package documentation
