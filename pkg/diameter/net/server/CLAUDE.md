# CLAUDE.md - server package

This file provides guidance for working with the Diameter server package (`pkg/diameter/net/server`).

## Package Overview

The `server` package implements a dual-transport Diameter server that simultaneously listens on both SCTP and TCP ports. It uses a semaphore-based worker pool to handle concurrent client connections and provides structured logging for diagnostics.

## Key Types

### Server

The main server struct that manages Diameter peers and connections:

**Fields:**
- `wp` - `*WorkerPool` - Manages concurrent workers
- `sctpListener` - `transport.SctpListener` - SCTP listener
- `tcpListener` - `transport.TcpListener` - TCP listener
- `state` - `atomic.Int32` - Current server state (Stopped/Running/Shutdown)
- `verbLevel` - `atomic.Int32` - Current verbosity level (Quiet/Error/Warn/Info/Debug)
- `autoReply` - `bool` - Whether to auto-reply to incoming messages
- `mu` - `sync.Mutex` - Mutex for configuration access
- `ctx/cancel` - Context for listener control
- `wgLstr` - `sync.WaitGroup` - Tracks active listeners
- `env` - `*diameter.Diameter` - Diameter environment (for message processing)

**States:**
- `StateStopped` (0) - Initial/terminal state
- `StateRunning` (1) - Actively accepting connections
- `StateShutdown` (2) - Shutdown in progress

**Verbosity Levels:**
- `Quiet` (0) - No output
- `Error` (1) - Errors only
- `Warn` (2) - Warnings and above
- `Info` (3) - Informational and above
- `Debug` (4) - All messages

### WorkerPool

Semaphore-based worker pool for connection handling:

**Fields:**
- `maxWorkers` - Maximum concurrent workers (103 total)
- `workers` - `chan struct{}` - Semaphore channel
- `wg` - `sync.WaitGroup` - Tracks pending tasks
- `mu` - `sync.Mutex` - Mutex for wait group

**Max Distribution:** 100 clients + 2 listeners + 1 shutdown handler = 103

## Key Methods

### Server Methods

| Method | Description |
|--------|-------------|
| `New(env *diameter.Diameter) *Server` | Create new server with given environment |
| `Start(addr string, autoReply bool) error` | Start listening on address |
| `Wait()` | Block until all workers complete |
| `Shutdown()` | Graceful shutdown: disconnect peers, close listeners |
| `State() int32` | Get current state |
| `SetState(state int32)` | Atomically set state |
| `IsRunning() bool` | Returns true if state == Running |
| `IsStopped() bool` | Returns true if state == Stopped |
| `VerboseLevel() int32` | Get current verbosity level |
| `SetVerboseLevel(level int32)` | Set verbosity level |
| `Verbose(level int32, msg string, attrs ...slog.Attr)` | Log with level check |
| `Autoreply() bool` | Get auto-reply mode |
| `SetAutoreply(onOff bool)` | Set auto-reply mode |
| `Dump()` | Print server status to stdout |
| `context() context.Context` | Get listener control context |
| `envContext() context.Context` | Get environment context |

### WorkerPool Methods

| Method | Description |
|--------|-------------|
| `Execute(task func()) bool` | Queue task; returns false if at capacity |
| `Wait()` | Block until all tasks complete |
| `Add(delta int)` | Increment wait group counter |
| `Done()` | Decrement wait group counter |

## Usage Examples

### Basic Server Setup
```go
// Create Diameter environment
diam := diameter.New(nil)

// Create server
ctrlChan := make(chan int)
srv := server.New(diam)

// Start listening
err := srv.Start(":3868", true)  // autoReply = true
if err != nil {
    log.Fatal(err)
}

// Wait for shutdown
srv.Wait()
```

### Graceful Shutdown
```go
// External shutdown
go func() {
    // ... some condition ...
    srv.Shutdown()
}()

// Or from Diameter environment context
// Server automatically shuts down when env.Context() is cancelled
```

### Verbose Logging
```go
// Set debug level
srv.SetVerboseLevel(server.Debug)

// Log with structured attributes
srv.Verbose(server.Info, "Connection established",
    slog.String("peer", peerName),
    slog.String("address", addr))

// Log error
srv.Verbose(server.Error, "Failed to accept connection",
    slog.Any("error", err))
```

### Server Status
```go
srv.Dump()  // Output:
            // Server is RUNNING
            // Auto reply mode is: ON
            // Listening on:
            //   sctp://:3868
            //   tcp://:3868
```

## Server Lifecycle

### Start Sequence
1. Check if already running (return early if so)
2. Create listeners (SCTP + TCP)
3. For each listener that succeeds:
   - Call `runListener()` - starts accepting goroutine
   - Add to wait group
4. Start environment context watcher (triggers shutdown)
5. Set state to Running

### Accept Loop
```go
for {
    select {
    case <-envContext.Done():
        cancel()
    case <-context().Done():
        return
    default:
        conn, err := listener.Accept()
        if err != nil {
            // Handle errors (net.ErrClosed, syscall.EINVAL)
            continue
        }
        if !wp.Execute(func() {
            connHandler(conn)
        }) {
            // Pool exhausted
            srv.Verbose(server.Warn, "Number of clients exceeded")
            tr.Close()
        }
    }
}
```

### Connection Handler
```go
func (s *Server) connHandler(tr transport.ITransport) {
    // Create peer in environment
    peer, _ := s.env.NewPeerEx(tr, s.reply)
    
    // Wait for peer to close
    for peer.IsOpen() {
        select {
        case <-s.envContext().Done():
            s.cancel()
        case <-s.context().Done():
            return
        case <-peer.Context().Done():
            return
        }
    }
}
```

### Auto-Reply
When `autoReply` is enabled, the server automatically responds to Diameter messages:
1. Receives message via `connHandler`
2. Calls `s.reply(data, peer)` callback
3. If enabled: creates response and sends it
4. Returns `true` if reply was sent

## Shutdown Sequence

1. Atomically switch state from Running to Stopped
2. Log shutdown message
3. Disable auto-reply
4. Cancel context (signals all listeners)
5. Close SCTP listener
6. Close TCP listener
7. Wait for listener goroutines to complete (`wgLstr.Wait()`)
8. Disconnect all peers (clients only by default)
9. Wait for all workers to complete
10. Signal completion via environment wait group

## Error Handling

### Accept Loop Errors
- `net.ErrClosed` - Listener closed, exit goroutine
- `syscall.EINVAL` - Invalid argument (listener closed), exit goroutine
- Other errors - Logged and continue loop

### Connection Rejection
If worker pool is at capacity (`maxClients = 100`):
- Connection is rejected
- Warning is logged
- Transport is closed

## Constants

```go
// Maximum concurrent clients
const maxClients = 100

// Maximum workers (clients + listeners + handler)
const maxWorkers = maxClients + 2 + 1  // 103

// Server states
const (
    StateStopped   = 0  // Initial/terminal
    StateRunning   = 1  // Accepting connections
    StateShutdown  = 2  // In progress
)

// Server commands (for control channel)
const (
    Shutdown         = 0  // Initiate shutdown
    ShutdownComplete = 1  // Signal completion
)

// Verbosity levels
const (
    Quiet = 0  // No output
    Error = 1  // Errors only
    Warn  = 2  // Warnings and above
    Info  = 3  // Informational and above
    Debug = 4  // All messages
)
```

## Dependencies

- `tgdp/pkg/diameter` - Diameter environment
- `tgdp/pkg/diameter/net/transport` - Transport abstraction
- `tgdp/pkg/diameter/diwe` - Error types
- `golang.org/x/sys/unix` - System constants
- `github.com/ishidawataru/sctp` - SCTP support

## Notes

- **Dual Transport**: Server runs both SCTP and TCP listeners; one failing doesn't affect the other
- **Auto-Reply**: When enabled, all received messages trigger automatic responses
- **Client Management**: Server creates peers in the Diameter environment; cleanup is handled on shutdown
- **Goroutines**: All worker goroutines are managed via `WorkerPool.Execute()`
