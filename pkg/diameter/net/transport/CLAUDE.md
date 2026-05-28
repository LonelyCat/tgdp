# CLAUDE.md - transport package

This file provides guidance for working with the Diameter transport package (`pkg/diameter/net/transport`).

## Package Overview

The `transport` package provides the network transport abstraction layer for Diameter protocol communication. It supports both TCP and SCTP transports through a unified interface (`ITransport`) and provides listener interfaces for server implementations.

## Directory Structure

```
transport/
├── transport.go      # ITransport and IListener interfaces
├── tcp.go            # TCP transport implementation
├── sctp.go           # SCTP transport (common code)
├── sctp_linux.go     # Linux-specific SCTP (full implementation)
├── sctp_darwin.go    # macOS SCTP (not implemented)
└── error.go          # Transport error types
```

## Key Interfaces

### ITransport

Abstract interface for network transports (TCP or SCTP):

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
    Type() int  // TransportSctp or TransportTcp
}
```

### IListener

Interface for transport listeners used by the server:

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

## Transport Implementations

### TCP Transport (`*Tcp`)

| Method | Description |
|--------|-------------|
| `Connect(local, remote)` | Dial TCP connection using `net.DialTCP` |
| `Send(data)` | Write to TCP connection |
| `Recv()` | Read from TCP connection into fixed 64KB buffer |
| `IsConnected()` | Check if `Connection != nil` |
| `LocalAddr()`, `RemoteAddr()` | String representation of addresses |
| `LocalIp()`, `RemoteIp()` | `netip.Addr` versions |
| `LocalPort()`, `RemotePort()` | Port numbers |
| `Close()` | Close TCP connection |

**Listener Methods:**
- `Create(addr)` - Start TCP listener with `net.ListenTCP`
- `Accept()` - Accept TCP connection via `AcceptTCP`
- `Ready()` - Returns `listener != nil`
- `Uri()` - Returns `"tcp://host:port"`

### SCTP Transport (`*Sctp`)

| Method | Description |
|--------|-------------|
| `Connect(local, remote)` | Dial SCTP using `sctp.DialSCTP` (Linux only) |
| `Send(data)` | Write to SCTP connection |
| `Recv()` | Read from SCTP connection (64KB buffer) |
| `IsConnected()` | Check if `Connection != nil` |
| `LocalAddr()`, `RemoteAddr()` | String representation |
| `LocalIp()`, `RemoteIp()` | `netip.Addr` versions |
| `LocalPort()`, `RemotePort()` | Port numbers |
| `Close()` | Close SCTP connection |

**Listener Methods:**
- `Create(addr)` - Start SCTP listener (Linux: `sctp.ListenSCTP`)
- `Accept()` - Accept SCTP connection via `AcceptSCTP`
- `Ready()` - Returns `listener != nil`
- `Uri()` - Returns `"sctp://host:port"`

## Platform Support

| Feature | Linux | macOS |
|---------|-------|-------|
| SCTP Client | Full implementation | Returns `ErrNotImplemented` |
| SCTP Server | Full implementation | Returns `ErrNotImplemented` |
| TCP Client | Full | Full |
| TCP Server | Full | Full |

**Note:** SCTP is not supported on macOS. The `sctp_darwin.go` file returns `ErrNotImplemented` for all operations.

## Constants

```go
const (
    TransportTcp = 2
    TransportSctp = 1
)

const (
    DefaultTimeout   = 10   // seconds
    DefaultPort      = 3868 // Diameter default
    DefaultProtocol  = "sctp"
)
```

## Constructor Functions

### New(proto string) (ITransport, error)

Creates a transport instance based on the protocol:

```go
tr, err := transport.New("tcp")   // Returns &Tcp{}
tr, err := transport.New("sctp")  // Returns &Sctp{}
```

**Errors:**
- `&ErrUnknownProto{Proto: proto}` - Unknown protocol

### Error Types

```go
type ErrProtoUnsupported struct {
    Proto string
    Os    string
}
func (e *ErrProtoUnsupported) Error() string

type ErrUnknownProto struct {
    Proto string
}
func (e *ErrUnknownProto) Error() string
```

## Usage Examples

### Creating a Client Transport
```go
// TCP
tcpTr := &transport.Tcp{}
err := tcpTr.Connect(
    netip.MustParseAddr("192.168.1.1"),  // remote IP
    3868,                                 // remote port
    netip.MustParseAddr("192.168.1.2"),  // local IP
    0,                                     // local port (0 = OS assigned)
)

// SCTP (Linux only)
sctpTr := &transport.Sctp{}
err := sctpTr.Connect(
    netip.MustParseAddr("192.168.1.1"),
    3868,
    netip.MustParseAddr("192.168.1.2"),
    0,
)
```

### Creating a Server Listener
```go
// TCP
tcpListener := &transport.TcpListener{}
err := tcpListener.Create(":3868")
tr, err := tcpListener.Accept()

// SCTP
sctpListener := &transport.SctpListener{}
err := sctpListener.Create(":3868")
tr, err := sctpListener.Accept()
```

### Using the Factory Function
```go
func createTransport(proto string) (transport.ITransport, error) {
    tr, err := transport.New(proto)
    if err != nil {
        return nil, err
    }
    
    err = tr.Connect(remoteIP, 3868, localIP, 0)
    return tr, err
}
```

### Working with Both Transports
```go
func handleTransport(tr transport.ITransport) {
    fmt.Printf("Transport: %s\n", tr.Name())
    fmt.Printf("Type: %d\n", tr.Type())
    fmt.Printf("Connected: %v\n", tr.IsConnected())
    fmt.Printf("Remote: %s\n", tr.RemoteAddr())
    
    // Type-specific operations (if needed)
    if tr.Type() == transport.TransportSctp {
        // SCTP-specific logic
    }
}
```

## Implementation Notes

### Buffer Size
Both TCP and SCTP implementations use a fixed 64KB buffer (`var buf [65535]byte`) for receiving data. This is sufficient for Diameter messages (typically < 1KB) but may need adjustment for custom use cases.

### Connection State
- **TCP**: Connection state is tracked via `*net.TCPConn` pointer
- **SCTP**: Connection state is tracked via `*sctp.SCTPConn` pointer

### Timeout Handling
Timeouts are implemented using `Connection.SetDeadline()` with `time.Now().Add(duration)`.

### Error Storage
Both transports store errors in an `Err` field for later retrieval via `Error()` method.

## Dependencies

- `net` and `net/netip` - Standard library networking
- `github.com/ishidawataru/sctp` - SCTP support (Linux)
- `golang.org/x/sys/unix` - System constants for SCTP

## Testing Considerations

- Test with both TCP and SCTP where available
- Verify error handling when transport fails to connect
- Test timeout behavior
- Test concurrent send/recv operations
