
# Welcome to TGDP

Traffic Generator for Diameter Protocol (TGDP) is a specialized tool designed for generating Diameter test traffic. TGDP enables the creation of Diameter messages, traffic parameter configuration, testing, and simulation of interaction scenarios with Diameter servers using the AAA (Authentication, Authorization, Accounting) protocol, which is used in modern telecommunications networks to ensure security and access control.

TGDP includes the following components:

- Packet Processor – the central component of packet processing, ensuring the formation of correct data;
- Definition of the Diameter protocol written in Apple (Pickle) Pkl language;
- Configuration files (Appendix A) that allow you to customize operating parameters, written in YAML;
- CLI, REPL and scripting interfaces.

The use of TGDP and the Diameter protocol together ensures the reliability and security of the telecommunications network and provides the following capabilities:

- creation and analysis of data packets;
- testing and debugging of telecommunication networks.

TGDP operates in one of three modes:

- Basic (CLI) – used to send a series of Diameter requests from one application to one remote peer using the command line;
- Extended (REPL) – used to interactively execute commands and make changes to messages;
- Scripting (Lua) – used to run scripts written in the Lua language.


## Installation

TGDP fully support Linux OS (SCTP protocol must be activated) and partial MacOS (due MacOS not supports SCTP).

TGDP is distributed as part of the delivery kit:

- **tgdp** – TGDP executable file;
- **pkl** – executable file of the Pkl language, required for loading configurations;
- **configs** – folder with configuration files and description of the Diameter protocol dictionary in `Pkl` format.

**Note**: By default, configuration files are located in a subdirectory `.tgdp` of the user's home directory.

To install TGDP, you need to do the following steps:

- copy executable files to any directory from the PATH environment variable
- create folder `~/.tgdp`
- place the `configs` folder from the TGDP distribution kit into the `.tgdp` folder

Also TGDP can be build from sources.
Requirements:
- Go lang version 1.23 or above
- GNU make
- Apple Pkl compiler
- Library PCAP development

1. Install Go lang - https://go.dev
2. Install Pkl compiler - https://pkl-lang.org/main/current/pkl-cli/
3. Install `GNU make` and `library PCAP development` according to your OS
4. Build `tgdp`:
* Clone the repository from `github`:
```bash
git clone https://github.com/LonelyCat/tgdp/tgdp.git
```
* Make release binary:
```bash
make release
```
  - Copy `tgdb` binary to a directory in PATH (e.g. ~/.local/bin)
  - Run `tgdp -h` to ensure everything is OK


## Diameter Message Construction Rules
Diameter messages consist of AVPs.
AVPs can be mandatory or optional.

The rules are:
* If the AVP is _mandatory_ its value[s]  are _must be specified_ in the file `avp-data.yaml`.
* If the AVP is optional and its value[s] ​​are _specified_ in the `avp-data.yaml`, the AVP is _included_ in the message.
* If _no value is specified_, the AVP is _not included_ in the message.
* If an AVP has _multiple_ values, it will be included in the message _multiple times_ as a separate AVP.

##  Operating modes

The TGDP provides the following operating modes:
- Basic - CLI
- Extended - REPL
- Simple Diameter server
- Scripting with Lua language

### Basic mode (CLI)

To run the TGDP tool, you need to run the following command line:
```bash
tgdp [flags] <peer> <app> <command> [<command> ...]
```

Where:
-   **peer** – node name (specified in the `peers.yaml` configuration file);
-   **app** – the name or identifier of the “Diameter” application;
-   **command** – short name or numeric code of the command.

**Note** – string parameters are not case-sensitive (for example, the expressions `S6a` and `s6A` are equivalents).

The `peer` parameter defines the name of the remote node in the `peers.yaml` configuration file and is specified by a user-defined string. A description of the `peers.yaml` file is provided in Appendix A.

The `app` parameter is specified by a string name or a numeric identifier, according to the set of standards for mobile communications described in the 3GPP specification and integrated into the TGPD for testing network solutions in the field of mobile communications.

The `command` parameter is specified by a short name defined in the Pkl files or a digital code according to the 3GPP specification (for example, `UL` or `316` for `Update-Location`.

If more than one command is specified in the parameters, all of them will be transmitted to the remote node in the order in which they are sent.

In basic mode, the message structure is determined by the description provided in the Pkl files.

**Note**: AVP values are specified in the configuration file `avp-data.yaml`.

If an AVP is mandatory and AVP value is not specified in `avp-data.yaml`, TGDP fails with error and performs appropriate diagnostics.

TGDP supports the following switch structures and command line arguments (flags):

- -a – append data to an existing pcap file;
- -c string – specifies the path to the configuration files directory (by default “~/.tgdp/configs”);
- -d - dump kmown application id and commands and exit;
- -h – display help;
- -n – not sending requests to remote nodes (offline mode);
- -s address:port - run simple Diameter server
- -v int – displays information on the screen, where values from 0 to 3 correspond to the level of detail (default is 1). For more details see command `verbose` in REPL mode.
- -w string – write a message to a pcap file;
- -y – check the Diameter dictionary (description in Pkl).

#### Examples of using TGDP

To send an `S6a` application Update-Location (`UL`) message to the `hss` node, you should run the following command:
```bash
tgdp hss s6a ul
```

To send the Update Location (`UL`) and Purge-UE (`321`) messages of the S6a application to the `hss` node,  run the following command:
```bash
tgdp hss s6a ul 321
```

To send the Insert-Subscriber-Data (`ID`) of the application `S6a` to the node `mme`, with the recording of the request and response in the file `output.pcap` you need to execute the following command:
```bash
tgdp -w output.pcap mme s6a id
```

To create a query, use Insert-Subscriber-Data (`ID`) of the application `S6a` without sending it to the remote node, with recording the request and response in the file « output . pcap » you need to execute the following command:
```bash
tgdp -n -v 2 -w output.pcap mme s6a id
```

**Note**: in offline mode (-n flag) the "peer" parameter *must* be specified and described in the configuration, but can have fake parameters.

To check the Diameter description (Pkl files) for errors, run the following command:
```bash
tgdp -y
```
Note : The verification protocol is displayed on the screen. All other parameters, are ignored.

### Extended/advanced mode (REPL)

When working in advanced mode, the user interacts with TGDP using commands. The command's results are displayed on the screen.

To use the advanced mode, launch TGDP without specifying any parameters. The tool's readiness is indicated by the **`D>`** prompt displayed on the screen.

During startup commands are read and executed from the `~/.tgdp/autorun` file.
The command line is used in a one-line, one-command format. Comments are marked with the "#" symbol at the beginning of the line.
The command execution history is saved in the file `~/.tgdp/history`.

The list of commands and their descriptions are given in the table .

| Command | Synonyms | Description |
|--|--|--|
| help	| ?	| Help output |
| quit | exit bye | Exiting TGDP |
| batch | bat | Executing a command from a file |
| echo | | Sending text to print |
| list | | Viewing the "peer" and "AVP" lists |
| connect | | Connection to a remote node (peer) |
| close | | Close connection with a peer |
| send | | Send a message to a peer |
| receive | recv | Receive a message from a peer |
| run | | Execute a Lua script |
| server | | Run a local server |
| avp	| | Setting up and retrieving AVP data |
| verbose	| | Setting the output verbosity level |

**Note**: Use the TAB key to complete commands and [possible] parameters.

#### The `help` command
The `help` command is used to display information about commands on the screen.
If no argument is provided, a list of all available commands will be displayed.
```tgdp
D> help [command]
```

```tgdp
D> help
Available commands:
help - Display help information
quit - Quit from TGDP
batch - Execute commands from a batch file[s]
list - Show list of objects
connect - Connect to a peer
close - Close connection to a peer
echo - Print a text message
send - Send a message[s] to a peer
receive - Receive a message from a peer
run - Run a Lua script
server - Control Diameter server
verbose - Set verbosity level
avp - Manage AVPs & global values
```

```tgdp
D> help connect
Connect to a peer
Usage: connect peer | address [port] [transport]
```

#### The `batch` command
The `batch` command is used to execute commands from files.
The list of files to execute is defined using the `file` parameter.
```tgdp
D> batch <file> [file ...]
```

```tgdp
D> batch set-org-realm.tgdp
```
**Note**: По умолчанию TGDP ищет файл в каталоге `~/.tgdp/batch`.

#### The "echo" command
The `echo` command is used to display a text to the terminal.
To print arbitrary text, use the `text` parameter.
```tgdp
D> echo [text]
```

To print empty line use `echo` without a parameters.
```tgdp
D> echo Hello World !
Hello World!
```

####  The `list` command
The `list` command is used to view lists of `peers` or `AVP` values loaded from configuration files.
```tgdp
D> list <avps | peers>
```
The `avps` parameter is used to get a list of AVP values.
The `peers` parameter is used to get a list of remote peer nodes.
**Note** : The `*` marker is used to indicate an active connection to a node.
```tgdp
D> list peers
0 * mme1 172.19.101.155 3868 SCTP
1   pcrf1 192.168.23.11 3871 TCP
2   test localhost 3868 TCP
3 * dra1 172.19.101.201 3868 SCTP
4   dra2 172.19.101.202 3868 SCTP
```

```tgdp
D> list avps
Auth-Application-Id (258) = 16777251
Auth-Application-Id (258) = 16777217
Auth-Application-Id (258) = 16777312
Auth-Session-State (277) = 1
Cancellation-Type (1420) = 2
Destination-Host (293) = hss01.vn.mnc040.mcc250.3gppnetwork.org
Destination-Realm (283) = vn.mnc040.mcc250.3gppnetwork.org
```

#### The `connect` command
The `connect` command is used to connect to a remote node (peer).
The following parameters are used to apply the command:
```tgdp
D> connect <peer> | <id> | <address [port] [transport]>
```
* `peer` – the name of a node from the list of known ones
* `id` – node number from the result of `list peers`  command
* `address` – IP address of peer
* `port` – network port (default 3868)
* `transport` –  transport protocol SCTP (by default) or TCP

**Note**: If an address was used for the connection, the node is assigned a name of the following form: `peer-<address>:<port>`.

```tgdp
D > connect mme1
```

```tgdp
D> connect 172.19.101.201 3868 sctp
D> list peers
0 * mme1 172.19.101.155 3868 SCTP
1   hss1 198.18.11.99 3868 TCP
2 * peer-172.19.101.201:3868 172.19.101.201 3868 SCTP
```

```tgdp
D> connect 1
```

#### The `close` command
The `close` command is used to close the connection to a peer.
The following parameters are used to apply the command:
```tgdp
D> close <peer | id>
```
* `peer` – the name of a node from the list of known ones
* `id` – node number from the result of `list peers`  command

```tgdp
D> close mme1
```
```tgdp
D> close 3
```

#### The `send` command
The `send` command is used to send a given message to a connected node.
The following parameters are used to apply the command:
```tgdp
D> send [flags] <type> <peer> <app> <message> [message ...]
```
`type` – `[req]uest` or `[ans]wer`
`peer` – name or id the same as in command `connect`
`app` – Application ID ( name or code )
`message` – message (short name or code).

The following flags are used to apply the command:
`-w | --wait` – wait for a response message[s]

**Note**: if `wait` flag is specified TGDP will be wait for an message to infinity. Press `^C`to interrupt waiting.

```tgdp
send –w req dra1 s6a ul
```
```tgdp
send answer mme1 0 dw
```

#### The `receive` command
The `receive` command is used to receive messages from a connected peer.
The following parameters are used to apply the command:
```tgdp
D> receive [flags] <peer>
```
`peer` – name or id the same as in command `connect`

The following flags are used to apply the command:
`-w | --wait` – wait for message to be received

**Note**: if `wait` flag is specified TGDP will be wait for an message to infinity. Press `^C`to interrupt waiting. If flags not specified and no messages to receive - command returns immediately.

```tgdp
receive –w dra1
```
```tgdp
recv 1
```

#### The `avp` command
The `avp` command is used to manage AVP data.
The following parameters are used to apply the command:
```tgdp
avp <action> <avp> [index] [value]
```
* `action` - info | get | set | add | delete
  - `info` – get information about AVP
  - `get` – get AVP value
  - `set` – set AVP value
  - `add` – add an AVP value
  - `delete` – delete an AVP value
  - `load` – load AVP value[s] from YAML file
* `avp` – id or name
* `index` – index of the value
* `value` – the value of the AVP

For grouped type AVP `avp info` command shows members of the group includding the foloowing attributes:
* `*` - means requred member
* `<number>` - means the maximum number of entries allowed in a message

```tgdp
D> avp info Auth-Application-Id
Code: 258
Name: Auth-Application-Id
Type: Unsigned32
```

```tgdp
D> avp info Vendor-Specific-Application-Id
Code: 260
Name: Vendor-Specific-Application-Id
Type: Grouped
Members:
*   1 Vendor-Id
    1 Auth-Application-Id
    1 Acct-Application-Id
```
**Note**: asterisks '*' mark required members

```tgdp
D> avp get 258
0 - Auth-Application-Id (258) = 16777251
1 - Auth-Application-Id (258) = 16777217
2 - Auth-Application-Id (258) = 16777312
```

```tgdp
D> avp add Auth-Application-Id 4
0 - Auth-Application-Id (258) = 16777251
1 - Auth-Application-Id (258) = 16777217
2 - Auth-Application-Id (258) = 16777312
3 - Auth-Application-Id (258) = 4
```

```tgdp
D> avp delete Auth-Application-Id 2
0 - Auth-Application-Id (258) = 16777251
1 - Auth-Application-Id (258) = 16777217
2 - Auth-Application-Id (258) = 4
```
**Note**: if index is not specified, all values is deleted.

```tgdp
D> avp set Auth-Application-Id 2 16777312
0 - Auth-Application-Id (258) = 16777251
1 - Auth-Application-Id (258) = 16777217
2 - Auth-Application-Id (258) = 16777312
```

```tgdp
D> avp load upd_loc.yaml
```

**Note**: The YAML content should be similar to `avp-data.yaml`.

**Note**: By default TGDP looks up file in `~/.tgdp/yaml` directory.

#### The `run` command
The `run` command is used to execute a Lua script.
```tgdp
run <script.lua> [parameters]
```
```tgdp
run scripts/demo.lua
```
```tgdp
run script.lua IMSI 123450123456789
```

#### The `verbose` command
The `verbose` command is used to control the level of detail in the output.
The following parameters are used to apply the command:
```tgdp
D> verbose [level]
```
`level` – number from 0 to 3;
0 – do not output anything (silent mode)
1 – display Diameter messages
2 – display Diameter messages and data about partner nodes
3 – includes 2 and Common Messages (App ID = 0) messages

**Note**:  if parameter `level` is omitted - current level will be displayed.

```tgdp
D> verbose
Verbosity level: 3
```
```tgdp
D> verbose 1
Verbosity level: 1
```

### Simple server mode
TGDP can operate as a simple Diameter server. In this mode, TGDP simply attempts to respond to requests if the required AVPs are specified in the configuration.

The `server` can be run by two ways:
1. Using flags `-s` in CLI mode
2. Run command `server` in REPL mode

#### CLI mode
```bash
tgdp -s address:port
```
* `address` - IP address listen
* `port` - IP port port

**Note**: TGDP tries to listen both SCTP and TCP transport protocols.

```bash
$ tgdp -s localhost:3868

Server is running
Listening on:
  sctp://127.0.0.1:3868
  tcp://127.0.0.1:3868
```
To shutdown - press `^C`.

#### REPL  mode
To handling server in REPL mode use the command `server`.
```tgdp
D> server <action> <address> [port]
```

* `action` - start | stop | status
  - `start` – start the server
  - `stop` – stop the server
  - `status` – request server status
* `address` – listening IP address
* `port` – listening port (default 3868)

```tgdp
D > server start localhost
Listening on sctp://127.0.0.1:3868
Listening on tcp://127.0.0.1:3868
```
```tgdp
D> server status
The server is running
```
```tgdp
D> server stop
Received shutdown signal, gracefully stopping...
TCP listener stopped
SCTP listener stopped
```
**Note**: in REPL mode server is background process, so user can continue interact with TGDP.

### Scripting in Lua
TGDP provides virtual module for Lua language.
Virtual means TGDP executes Lua scripts in its environment.

A description of Lua API provided by TGDP is given in file `Lua-scripting.md`.

To run Lua scripts via TGDP :
```bash
tgdp [-c <path>] @<script.lua> [parameters]
```
or
```tgdp
D> run <script.lua> [parameters]
```

## Appendix A - configuration files
The configuration files are:
* `avp-data.yaml` - AVPs data
* `peers.yaml` - remote peers definitions

By default, configuration files are located in the `~/.tgdp/configs` folder.
To specify an alternative path, use the command line switch: `-c <path>`.

### A.1 - avp-data.yaml
Format:
```yaml
<AVP-Name> : <value>
```
`AVP-Name` - name of an AVP according to RFC or 3GPP specification.
`value` - AVP value.

**Note**: AVP names are case insensitive.

Quotes `"` are generally not required for `value`, but are recommended for string AVP types and require if string consists of numbers.

```yaml
# Identity type - not require quotes
Origin-Host: host.realm.org

# UTF8String type - require quotes
User-Name: "123450123456789"

# Integer32 type
Vendor-Id: 10415
```

The value for `ENUMERATED` AVP type can be defined as either an integer or a mnemonic name representation.
The correspondence between the mnemonic name and the numeric code is written in Pkl files.
```yaml
Cancelation-Type: SUBSCRIPTION_WITHDRAWWAL
```
```yaml
# SUBSCRIPTION_WITHDRAWWAL = 2
Cancelation-Type: 2
```
For the `GROUPED` AVP type, only the name is specified without a value. Next, at an offset, are the group members with values that define the composition of the group AVP.
```yaml
# Grouped AVP
Vendor-Specific-Application-Id:
# Group members
  Vendor-Id: 10415
  Auth-Application-Id: 16777251
```
If an AVP requires multiple values, it should be defined as an YAML array:
```yaml
Auth-Application-Id:
  - 16777251
  - 16777217
  - 16777312
```

### A.2 - peers.yaml
Each peer node is described by the following schema:
```yaml
<name>:
  address: <IP address>
  port: <SCTP | TCP port>
  protocol: <"SCTP" | "TCP">
```
* `name` – the name of the node specified by the user
* `address` – the IP address or domain name
* `port` – the port used to connect to the peer (3868 by default)
* `protocol` – the protocol used to communicate with the node SCTP (default) or TCP

```yaml
hss1:
  address: 192.168.1.111
  port: 3868

pcrf:
  address: pcrf.operator.org
  port: 3870
  protocol: tcp
```
