# CLAUDE.md - node package

This file provides guidance for working with the Diameter node/peer management package (`pkg/diameter/net/node`).

## Package Overview

The `node` package implements Diameter protocol peer management with support for both client and server connections. It handles connection state management, transport abstraction (TCP/SCTP), network routing, and message handling with auto-reply capabilities.

## Key Types

### Node

Represents a Diameter peer connection with full lifecycle management:

**Core Fields:**
- `Name` - Peer identifier
- `Address` - Remote IP address or hostname
- `RemotePort` - Remote TCP/SCTP port (default 3868)
- `LocalPort` - Local TCP/SCTP port (auto-assigned if 0)
- `Type` - Peer type (e.g., "MME", "HSS")
- `Timeout` - Connection timeout in seconds
- `RouteInfo` - Network routing details (local IP, gateway, interface)

**Internal Fields:**
- `tr` - Transport layer (`transport.ITransport`)
- `state` - Atomic connection state
- `client` - Boolean flag for client vs server
- `parent` - Reference to parent `Nodes` collection
- `ctx/cancel` - Context for cancellation
- `rxChan` - Channel for received data
- `ccChan` - Channel for interrupt signals
- `diaApi` - Diameter API interface for message processing
- `ucb` - User-defined callback function

### Nodes

Thread-safe collection for managing multiple peers:
- Uses `sync.RWMutex` for concurrent access
- Supports iteration via Go 1.23 `iter.Seq`
- Provides YAML-based configuration loading

### RouteInfo

Network routing details obtained via `go-netroute`:
- `IfaceId` - Network interface index
- `IfaceMac` - Interface MAC address
- `RemoteIp` - Remote IP address
- `LocalIp` - Local IP address used for connection
- `GwIp` - Gateway IP address

## State Machine

Nodes transition through these states (atomic int32 values):

| State | Value | Description |
|-------|-------|-------------|
| `StateClosed` | 0 | Initial state or closed connection |
| `StateWaitConnAck` | 1 | Waiting for connection acknowledgment |
| `StateWaitCEA` | 2 | Waiting for Capabilities-Exchange-Answer |
| `StateIOpen` | 3 | Connection open as initiator (client) |
| `StateROpen` | 4 | Connection open as responder (server) |
| `StateSuspect` | 5 | Peer is suspect/unresponsive |
| `StateReOpen` | 6 | Reconnecting after failure |
| `StateShuttingDown` | 7 | Connection shutdown in progress |
| `StateWaitData` | 8 | Waiting for data arrival (during recv) |

**Connection Flow:**
- **Client**: Closed → WaitConnAck → WaitCEA → IOpen
- **Server**: Closed → WaitCEA (on incoming connection) → ROpen

## Constants

### State Constants
```go
StateClosed       // 0
StateWaitConnAck  // 1
StateWaitCEA      // 2
StateIOpen        // 3
StateROpen        // 4
StateSuspect      // 5
StateReOpen       // 6
StateShuttingDown // 7
StateWaitData     // 8
```

### Transport Type Constants
```go
TransportUnknown = 0
TransportSctp    = 1
TransportTcp     = 2
```

### Configuration Keys
```go
ConfKeyName      = "name"
ConfKeyAddress   = "address"
ConfKeyPort      = "port"
ConfKeyTransport = "transport"
ConfKeyTimeout   = "timeout"
```

## Key Methods

### Node Methods

| Method | Description |
|--------|-------------|
| `Connect()` | Establish connection, send Capabilities-Exchange Request, wait for CEA |
| `Disconnect()` | Send Disconnect-Peer-Request, then close connection |
| `Close()` | Close transport, cleanup channels, remove from collection (if client) |
| `SendTo(data []byte)` | Send raw bytes to peer (calls `tr.Send()`) |
| `RecvFrom(wait bool)` | Receive with interrupt signal support; returns `[]byte, error` |
| `SetTimeout(timeout int)` | Apply timeout to transport layer |
| `Transport()` | Get transport interface |
| `IsOpen()` | Check if state is IOpen, ROpen, or WaitData |
| `IsClient()` | Returns true if this is a client connection |
| `IsClosed()` | Returns true if state is Closed |
| `Interrupt()` | Send interrupt signal to unblock RecvFrom |
| `Context()` | Get node's context for cancellation |
| `GetRouteInfo()` | Query routing table for local interface/gateway |
| `Lock/Unlock/TryLock` | Mutex operations |

### Nodes Collection Methods

| Method | Description |
|--------|-------------|
| `NewNodes()` | Create empty collection |
| `NewPeer(name, addr, port, proto, timeout, diaApi)` | Create server-side peer, add to collection |
| `NewPeerEx(tr, diaApi, ucb)` | Create client from accepted transport, auto-add to collection |
| `GetByName(name)` | Get peer by name (case-insensitive) |
| `Remove(name)` | Remove peer by name |
| `DisconnectAll(clients bool)` | Disconnect all or client-only peers |
| `InterruptAll()` | Send interrupt to all peers |
| `Iter()` / `Iter2()` | Iterator over peers |
| `LoadFromFile(yamlFile, diaApi)` | Load peers from YAML config |
| `Print()` / `Dump()` | Print all peers to stdout |

## Usage Examples

### Creating a Server-Side Peer
```go
diaApi := diameter.New(env)
nodes := node.NewNodes()

node, err := nodes.NewPeer(
    "peer1",           // name
    "192.168.1.1",     // address
    3868,              // port
    "sctp",            // transport protocol
    30,                // timeout
    diaApi,            // diameter API
)
if err != nil {
    // handle error
}
err = node.Connect()
```

### Creating a Client from Accepted Connection
```go
tr, err := listener.Accept()  // SCTP or TCP listener
if err != nil {
    // handle error
}

node, err := nodes.NewPeerEx(tr, diaApi, myCallback)
if err != nil {
    // handle error
}
// Node is already in Open state
```

### Using the Iterator (Go 1.23)
```go
for peer := range nodes.Iter() {
    if peer.IsOpen() {
        fmt.Printf("Peer %s is open\n", peer.Name)
    }
}
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
  port: 3868
  transport: tcp
  timeout: 20
```

### Custom Callback
```go
func myCallback(data []byte, peer *node.Node) bool {
    // Process message
    msg, _ := diaApi.BytesToMessage(data)
    
    // Auto-reply if needed
    if shouldReply(msg) {
        response := diaApi.CreateResponse(data)
        peer.SendTo(response)  // nolint: errcheck
        return true  // message handled, don't queue to rxChan
    }
    
    return false  // message not handled, queue to rxChan
}
```

## Auto-Reply Mechanism

The `Node` automatically handles common messages (Application ID == 0):
1. Receives message via `asyncHandler`
2. Checks if it's a common message via `diaApi.IsCommonMessage()`
3. If request: auto-generates response via `replyCommonMessage()`
4. If answer or non-common: queues to `rxChan`

User callbacks can override this behavior by returning `true`.

## Dependencies

- `tgdp/pkg/diameter` - Diameter environment and API
- `tgdp/pkg/diameter/net/transport` - TCP/SCTP transport abstraction
- `tgdp/pkg/diameter/diwe` - Custom error types
- `github.com/libp2p/go-netroute` - Network route lookup
- `golang.org/x/sys/unix` - System constants (Linux SCTP)

## Notes

- **Port Assignment**: If `LocalPort` is 0, a random port is assigned in range `[3868, 3868+32768]`
- **Client Cleanup**: Client nodes are automatically removed from collection on close
- **Thread Safety**: All public methods use mutex locking internally
- **Interrupt Handling**: `RecvFrom` can be interrupted via SIGINT (Ctrl+C)
