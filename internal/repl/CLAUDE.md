# REPL Package

Interactive Read-Eval-Print Loop for TGDP Diameter testing.

## Architecture

```
repl.go           # Core REPL loop, command registration, batch execution
error.go          # REPL error types

Subpackages:
  avp/            # AVP management commands (info, get, set, add, delete, load, purge)
  close/          # Close peer connection command
  comp/           # Tab-completion helpers for dynamic objects (peers, AVPs, apps, files)
  connect/        # Connect to peer command
  echo/           # Print text command
  list/           # List peers/AVPs commands
  receive/        # Receive message from peer command
  script/         # Run Lua script command
  send/           # Send request/answer message command
  server/         # Diameter server control (start/stop/status)
  verbose/        # Set verbosity level command
```

## Key Design Patterns

- Each subpackage implements `RootCommand` (*cobra.Command) and `CompList()` for tab-completion
- Commands access `*diameter.Diameter` via `cmd.Context().Value(diameter.EnvContext)`
- `list.PeerNameById()` resolves numeric peer IDs to names for cross-command use
- Batch files (`batch` command) execute command scripts from `config.BatchDir()`

## Commands Reference

| Command | Description |
|---------|-------------|
| `avp <subcmd>` | Manage AVP values (info, get, set, add, delete, load, purge) |
| `batch <file>` | Execute commands from batch file |
| `close <peer>` | Close connection to peer |
| `connect <peer\|addr>` | Connect to peer by name or address[:port] |
| `echo <text>` | Print text |
| `help [cmd]` | Show help |
| `list peers\|avps` | List peers or AVP values |
| `quit` | Exit REPL |
| `receive [-w] <peer>` | Receive message (wait with -w) |
| `run <script.lua>` | Execute Lua script |
| `send [-w] req\|ans <peer> <app> <msg>` | Send message |
| `server start\|stop\|status` | Control Diameter server |
| `verbose [level]` | Get/set verbosity |

## Dependencies

- `github.com/chzyer/readline` - Terminal readline with history and completion
- `github.com/spf13/cobra` - Command framework
- `tgdp/pkg/diameter` - Core Diameter protocol
- `tgdp/internal/config` - Configuration paths
- `tgdp/internal/cli` - CLI message send/receive logic
- `tgdp/internal/lua` - Lua scripting runtime