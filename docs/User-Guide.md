# TGDP User Guide

## Table of contents

- [Introduction](#introduction)
- [Quick Start](#quick-start)
  - [1. Installation](#1-installation)
  - [2. Configure a Peer](#2-configure-a-peer)
  - [3. Configure AVP Data](#3-configure-avp-data)
  - [4. Send a Message](#4-send-a-message)
- [Installation](#installation)
  - [Building from Source](#building-from-source)
- [Configuration](#configuration)
  - [Main Configuration (`config.yaml`)](#main-configuration-configyaml)
  - [Peers (`peers.yaml`)](#peers-peersyaml)
  - [AVP Data (`avps.yaml`)](#avp-data-avpsyaml)
- [Message Creation Rules](#message-creation-rules)
- [Operating Modes](#operating-modes)
  - [1. CLI Mode](#1-cli-mode)
  - [2. REPL Mode](#2-repl-mode)
  - [3. Server Mode](#3-server-mode)
  - [4. Lua Scripting](#4-lua-scripting)
- [REPL Command Reference](#repl-command-reference)
  - [Command `help`](#command-help)
  - [Command `quit`](#command-quit)
  - [Command `batch`](#command-batch)
  - [Command `echo`](#command-echo)
  - [Command `peer`](#command-peer)
  - [Command `send`](#command-send)
  - [Command `receive`](#command-receive)
  - [Command `run`](#command-run)
  - [Command `server`](#command-server)
  - [Command `avp`](#command-avp)
  - [Command `pcap`](#command-pcap)
  - [Command `verbose`](#command-verbose)

## Introduction

The Traffic Generator for Diameter Protocol (TGDP) is a specialized tool for creating, sending, and analyzing Diameter protocol traffic.
It allows you to build Diameter messages, configure traffic parameters, and simulate interactions with Diameter servers,
making it an essential tool for testing and debugging AAA (Authentication, Authorization, Accounting) functionalities in modern telecommunications networks.

**Key Features:**

* **Flexible Message Creation:** Construct complex Diameter messages using simple configuration files.
* **Multiple Operating Modes:** Use TGDP as a command-line tool (CLI), an interactive shell (REPL), a simple server, or via Lua scripting.
* **Rich Protocol Support:** Based on a Diameter protocol definition written in the Pkl language, allowing for easy extension and validation.
* **Easy Configuration:** Define peers and AVP (Attribute-Value Pair) data using straightforward YAML files.

---

## Quick Start

This section will guide you through installing TGDP, configuring a remote peer, and sending your first Diameter message.

### 1. Installation

First, ensure you have:
* `tgdp` executable
* `pkl` compiler executable
* `configs` directory
* `libpcap.so` installed

1. Copy the `tgdp` and `pkl` executable to a directory in your system's `PATH` (e.g., `/usr/local/bin` or `~/.local/bin`).
2. Create a configuration directory in your home folder:
```sh
mkdir -p ~/.tgdp
```
3. Copy the contents of the `configs` directory provided with TGDP into the newly created `~/.tgdp` folder. To match the default `config.yaml`, place the data files in a `data` subdirectory. Your final structure should look like this:
```
~/.tgdp/
├── config.yaml
├── data/
│   ├── avps.yaml
│   └── peers.yaml
└── pkl/
    └── Diameter.pkl
    └── dictionary.pkl
    └── apps.pkl
    └── avps.pkl
```

### 2. Configure a Peer

Define the remote Diameter peers you want to communicate with in `~/.tgdp/data/peers.yaml`.

For example, to define a peer naamed `hss-test`:

```yaml
# ~/.tgdp/data/peers.yaml
hss-test:
  address: 192.168.1.100
  port: 3868
  protocol: sctp
```

### 3. Configure AVP Data

Define the data for the AVPs that will be included in your message in `~/.tgdp/data/avps.yaml`.

```yaml
# ~/.tgdp/data/avps.yaml
Origin-Host: "client.example.org"
Origin-Realm: "example.org"
Destination-Realm: "test.net"
User-Name: "1234567890"
Vendor-Specific-Application-Id:
  Vendor-Id:            10415
  Auth-Application-Id:  16777251
```

### 4. Send a Message

Now, you can send a message from the command line.
This example sends an **Update-Location-Request (ULR)** for the **S6a** application to the `hss-test` peer.

```sh
tgdp hss-test s6a ul
```

TGDP will construct the message, send it to the configured peer, and print the response.

---

## Installation

**Warning**: Apple's Pkl compiler binary executable is required in PATH. For more detail visit [www.pkl-lang.org](www.pkl-lang.org).

TGDP is distributed as a pre-compiled binary. For manual installation:

1. **Copy Executable:** Place the `tgdp` binary in a directory included in your `PATH` environment variable (e.g., `~/.local/bin`, `/usr/local/bin`).
2. **Create Config Directory:** Create the default configuration directory: `mkdir -p ~/.tgdp`.
3. **Copy Config Files:** Place the `configs` folder from the TGDP distribution into `~/.tgdp`.

### Building from Source

**Requirements:**
* Go 1.25+
* PCAP development library (`libpcap-dev` on Debian/Ubuntu, `libpcap-devel` on CentOS/RHEL).
* GNU Make

**Steps:**
1. **Install Go:** [go.dev](https://go.dev)
2. **Install Pkl:** [pkl-lang.org](https://pkl-lang.org/main/current/pkl-cli/)
3. **Install Dependencies:** Use your system's package manager to install `make` and the `PCAP library`.
4. **Build:**
```sh
git clone https://github.com/LonelyCat/tgdp/tgdp.git
cd tgdp
make [release]
```
5. Copy/move the resulting `tgdp` binary to your `PATH`.

---

## Configuration

TGDP uses a central `config.yaml` file for configuration, located by default in `~/.tgdp/config.yaml`.
You can specify a different configuration directory using the `-c <path>` flag.

### Main Configuration (`config.yaml`)

This file defines the paths to other data files and general operating modes.

**Schema:**
```yaml
avps_data_file: "data/avps.yaml"    # Path to the global AVPs data file
peers_data_file: "data/peers.yaml" # Path to the peers data file
batch_subdir: "batch"              # Subdirectory for REPL mode batch files
yaml_subdir: "yaml"                # Subdirectory for REPL mode YAML files
diameter_mode: "transaction"       # Diameter mode - "transaction" or "session"
dictionary_file: "pkl/dictionary.pkl" # Path to the PKL Diameter dictionary data file
```

### Peers (`peers.yaml`)

This file defines the remote Diameter peers (servers) TGDP can connect to. It is located at the path specified by `peers_data_file` in `config.yaml`.

**Schema:**
```yaml
<peer-name>:
  address: <IP address or FQDN>
  port: <Network Port>
  protocol: <"sctp" | "tcp">
```
**`<peer-name>`**: A custom name for the peer.
**`address`**: The IP address or domain name of the peer.
**`port`** (optional): The target port. Defaults to `3868`.
**`protocol`** (optional): The transport protocol. Defaults to `sctp`.

**Example:**
```yaml
hss1:
  address: 192.168.1.111
  port: 3868
  protocol: sctp

pcrf:
  address: pcrf.operator.org
  port: 3870
  protocol: tcp
```

### AVP Data (`avps.yaml`)

This file provides the values for the AVPs used to construct Diameter messages. It is located at the path specified by `avps_data_file` in `config.yaml`.

**Format:**
```yaml
<AVP-Name>: <value>
```

**Rules:**
* AVP names are case-insensitive.
* Use quotes for string values, especially if they consist only of numbers.
* For `ENUMERATED` AVPs value can be defined by code or mnemonic name.
* For `GROUPED` AVPs, list member AVPs indented below the group name.
* To specify multiple values for an AVP, use a YAML array.

**Examples:**
```yaml
# String types
Origin-Host: tgdp.client.net
User-Name: "123450123456789"

# Integer type
Vendor-Id: 10415

# Enumerated type (can use mnemonic name or code (integer value))
Cancellation-Type: SUBSCRIPTION_WITHDRAWAL
# SUBSCRIPTION_WITHDRAWAL is equivalent to 2
# Cancellation-Type: 2

# Grouped AVP
Vendor-Specific-Application-Id:
  Vendor-Id: 10415
  Auth-Application-Id: 16777251

# AVP with multiple values
Auth-Application-Id:
  - 16777251
  - 16777217
```

---

## Message Creation Rules

TGDP constructs Diameter messages based on the command definition in the Pkl files and the AVP data provided in `avps.yaml`
(or loaded dynamically in REPL mode). The logic is as follows:

**Mandatory AVPs**: 
* If an AVP is defined as mandatory for a specific command, its value **must** be present in the AVP data.
* If it is missing, TGDP will report an error and fail to create the message.

**Optional AVPs**:
* If an optional AVP **has a value** defined in the AVP data, it will be **included** in the message.
* If an optional AVP **does not have a value** defined, it will be **omitted** from the message.

**Multiple Values**: If an AVP is configured with multiple values (using a YAML array), the AVP will be included in the message multiple times.

This system allows you to maintain a single `avps.yaml` file with all necessary AVPs and trust that TGDP will correctly build the message
with only the required and specified optional AVPs for any given command.

---

## Operating Modes

### 1. CLI Mode

The CLI mode is for sending one or more pre-configured Diameter messages directly from the shell.

**Usage:**
```sh
tgdp [flags] <peer> <app> <command> [<command> ...]
```
**`peer`**: The peer name defined in `peers.yaml`.
**`app`**: The Diameter application name or ID (e.g., `s6a`, `16777251`).
**`command`**: The command's short name or code (e.g., `ULR`, `316`).

**Common Flags:**
* `-a` – append data to an existing pcap file
* `-c <path>`: Path to the configuration directory
* `-d` - list known application id and commands and exit
* `-n`: Dry-run mode. Build the message but do not send it
* `-s <addr:port>`: Run in simple server mode
* `-v <level>`: Set verbosity level (0-3)
* `-w <file.pcap>`: Write the exchange to a PCAP file
* `-y`: Validate the Diameter dictionary (Pkl files) and exit

**Examples:**
```sh
# Send an S6a Update-Location-Request (ULR) to the 'hss1' peer
tgdp hss1 s6a ul

# Send ULR and Purge-UE-Request (PUR) and write to a pcap file
tgdp -w output.pcap hss1 s6a ul pu

# Send "Common Message" (0) Device-Watchdog (280)
tgdp -w output.pcap dra 0 280

# Validate Diameter dictionary
tgdp -y
```

### 2. REPL Mode

The interactive REPL (Read-Eval-Print Loop) is for dynamic message creation, connection management, and scripting.
To start it, run `tgdp` with no arguments.

You will be greeted with the `D>` prompt. Type `help` to see a list of commands.

**Key Features:**
* Commands are read from `~/.tgdp/autorun` on startup.
* Command history is saved in `~/.tgdp/history`.
* Tab completion is available for commands and parameters.

See the **REPL Command Reference** section below for a full list of commands.

### 3. Server Mode

TGDP can act as a simple Diameter server.

**From CLI:**

In this mode TGDP automatically replying to requests based on the data in `avps.yaml`.
```sh
# Listen on localhost:3868 for both SCTP and TCP
tgdp -s localhost:3868

Server is running
Listening on:
  sctp://127.0.0.1:3868
  tcp://127.0.0.1:3868
```

**From REPL:**
In this mode TGDP not automatically replying to requests and require user actions to `receive` request and `send` answer.
```tgdp-repl
D> server start localhost 3868
Server is running
Listening on:
  sctp://127.0.0.1:3868
  tcp://127.0.0.1:3868

D> server status
Server is running

D> server stop
Received shutdown signal, gracefully stopping...
SCTP listener stopped
TCP listener stopped
```

### 4. Lua Scripting

For complex scenarios, you can automate TGDP using Lua scripts. A detailed API reference is available in `Lua-scripting.md`.

**From CLI:**
```sh
# The '@' symbol tells TGDP to execute a script
tgdp @scripts/demo.lua IMSI 123450123456789
```

**From REPL:**
```tgdp-repl
D> run scripts/demo.lua IMSI 123450123456789
```

---

## REPL Command Reference

This section provides a detailed reference for all commands available in TGDP's interactive REPL mode.

 |  Command | Synonyms | Description  |
 | -- | -- | -- |
 |  help  | ?     | Help output  |
 |  quit | exit bye | Exiting TGDP  |
 |  batch | bat | Executing a command from a file  |
 |  echo |  |  Sending text to print  |
 |  peer |  |  Manage remote peers  |
 |  send |  |  Send a message to a peer  |
 |  receive | recv | Receive a message from a peer  |
 |  avp   |  |  Setting up and retrieving AVP data  |
 |  server |  |  Run a local server  |
 |  run |  |  Execute a Lua script  |
 |  pcap |  |  Save messages to a PCAP file  |
 |  verbose       |  |  Setting the output verbosity level  |

**Note**: Use the TAB key to complete commands and [possible] parameters.

### Command `help`
Displays a list of all commands or detailed help for a specific command.
**Alias:** `?`
**Usage:** `help [command]`
**Example:**
```tgdp-repl
D> help
Available commands:
  help - Display help information
  quit - Quit from TGDP
  batch - Execute commands from a batch file[s]
  avp - Manage AVP global values
  echo - Print a text message
  pcap - Save messages to a PCAP file
  peer - Manage remote peers
  receive - Receive a message from a peer
  run - Run a Lua script
  send - Send a message[s] to a peer
  server - Control Diameter server
  verbose - Set verbosity level

D> help send
Send a message[s] to a peer
Usage: send [flags] <type> <peer> <app> <message> [message ...]
```

### Command `quit`
Exits the REPL session.
**Alias:** `exit`, `bye`
**Usage:** `quit`

### Command `batch`
Executes a series of commands from a specified file.
**Alias:** `bat`
**Usage:** `batch <file1> [file2] ...`
By default TGDP looks up file in `~/.tgdp/batch` directory.
**Example:**
```tgdp-repl
D> batch setup-session.tgdp
```

### Command `echo`
Prints text to the console. Useful for scripting and adding comments to output.
**Usage:** `echo [text]`
**Example:**
```tgdp-repl
D> echo --- Starting Test Case 1 ---
--- Starting Test Case 1 ---
```

### Command `peer`
Manage remote peers.
**Usage:** `peer <list | info | open | close> <name> | <id> | <addess [port]>`
**Example:**
```tgdp-repl
D> peer list
D> peer open HSS
D> peer open 1.2.3.4 3868
D> peer close HSS
D> peer info HSS
```

### Command `send`
Constructs and sends a Diameter message to a connected peer.
**Usage:** `send <request [-w | --wait] | answer> <peer> <app> <message> [message ...]`
**Arguments:**
* `-w | --wait`: Wait for a response after sending.
* `type`: `req[uest]` for a request or `ans[swer]` for an answer.
* `peer`: The name or ID of the connected peer.
* `app`: The Application ID (name or code).
* `message`: The message name or code.

**Example:**
```tgdp-repl
D> send req -w hss1 s6a ul
```

### Command `receive`
Waits for and receives an incoming message from a peer.
**Alias:** `recv`
**Usage:** `receive [-w | --wait] <peer>`
**Arguments:**
`-w | --wait`: Wait indefinitely for a message. If omitted, the command returns immediately if no message is in the buffer.
**Example:**
```tgdp-repl
D> receive -w hss1
```

### Command `run`
Executes a Lua script.
**Usage:** `run <script.lua> [args ...]`
**Example:**
```tgdp-repl
D> run scripts/create-user.lua 1234567890
```

### Command `server`
Controls the built-in simple Diameter server.
**Usage:** `server <start | stop | status> [address] [port]`
**Example:**
```tgdp-repl
D> server start localhost 3868
D> server status
D> server stop
```

### Command `avp`
Manages AVP data in memory for the current session.
This is useful for dynamically changing AVP values without editing the `avps.yaml` file.

**Sub-commands:**
* `list`: Show AVPs with values.
* `info <avp>`: Show AVP definition.
* `get <avp>`: Display current value(s).
* `set <avp> <index> <value>`: Modify a value.
* `add <avp> <value>`: Add a new value.
* `delete <avp> [index]`: Delete one or all values.
* `load <file.yaml>`: Load AVP data from a file.
* `purge`: Clear all AVP data.

**Examples:**
```tgdp-repl
# Getting info about AVP Vendor-Specific-Application-Id
D> avp info Vendor-Specific-Application-Id
Code: 260
Name: Vendor-Specific-Application-Id
Type: Grouped
Members:
*   1 Vendor-Id
    1 Auth-Application-Id
    1 Acct-Application-Id

# Check the current User-Name
D> avp get User-Name
0 - User-Name (1) = "1234567890"

# Change the User-Name for the next message
D> avp set User-Name 0 "0987654321"
User-Name (1):
  0: 0987654321

# Delete value with index 2
D> avp delete Auth-Application-Id 2
0 - Auth-Application-Id (258) = 16777251
1 - Auth-Application-Id (258) = 16777217
```
**Notes**:
* for `info` asterisks '*' mark required members.
* if index is not specified for `delete`, all values is deleted.
* the YAML content should be similar to `avps.yaml`.
* by default TGDP looks up file in `~/.tgdp/yaml` directory.

### Command `pcap`
Manages saving data in PCAP files.
**Usage:** `pcap <open [-t] <file.pcap> | close | status>`
**Arguments:**
* `open`: Opens new PCAP file. If file exists and flag `-t` present - file will be `truncated`.
* `close`: Closes opened PCAP file.
* `status`:  Shows status of PCAP "session".

**Example:**
```tgdp-repl
D> pcap open trace.pcap
D> pcap status
PCAP status:
  state: OPEN
   mode: APPEND
   file: trace.pcap
D> pcap close
```

### Command `verbose`
Sets the verbosity level of the output.
**Usage:** `verbose [level]`
**Levels:**
* `0`: Silent mode.
* `1`: Display Diameter messages.
* `2`: Level 1 + peer connection data.
* `3`: Level 2 + common messages (e.g., DWR/DWA).
**Example:**
```tgdp-repl
D> verbose 3
D> verbose
Verbosity level: 3
```
