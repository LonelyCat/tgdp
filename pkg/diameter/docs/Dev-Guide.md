
# Diameter Package Developer Guide (draft)

The `diameter` package provides a complete implementation of the Diameter protocol (RFC 6733),
allowing developers to create, manipulate, serialize, and transmit Diameter messages.

## Core Concepts

### Diameter Environment (`Diameter`)
The `Diameter` struct is the central coordinator for all protocol operations. It manages:
- **Dictionary**: Metadata for Applications, Commands, and AVPs.
- **AvpStore**: A thread-safe repository of AVP values used for message templating.
- **Codecs**: Logic for converting between Go types and Diameter wire format.
- **Peers**: Management of network nodes and transport connections.
- **Context**: Lifecycle management for asynchronous operations.

### Messages (`Message`)
A `Message` represents a Diameter packet consisting of a 20-byte header and a sequence of AVPs.
- **Header**: Contains Version, Length, Flags, Command Code, Application ID, Hop-by-Hop and End-to-End identifiers.
- **Serialization**: Messages are serialized to big-endian wire format. Serialization is cached for performance.

### Attribute-Value Pairs (`Avp`)
`Avp` is the basic unit of data in Diameter. Each AVP consists of:
- **Header**: Code, Flags (V, M, P), and optional Vendor-ID.
- **Value**: Typed data encoded via a registered codec.
- **Alignment**: All AVPs are padded to a 4-byte boundary.

---

## API Reference

### Diameter Environment

#### `New(mode int32) (*Diameter, error)`
Creates and initializes a new Diameter environment.
- `ModeTransaction`: Appends timestamps to `Session-Id` for unique transactions.
- `ModeSession`: Uses static session identifiers.

#### `LoadDict(file string, format int) error`
Loads the Diameter dictionary from a file and registers all necessary codecs.

#### `LoadData(file string) error`
Loads the AVP data from a file with the specified append mode.

#### `LoadPeers(file string) error`
Loads peers configuration from a file.

#### `NewRequest(appId any, cmdCode any) (*Message, error)`
Creates a new request message. Automatically populates AVPs from the `AvpStore` based on dictionary rules.

#### `NewAnswer(appId any, cmdCode any) (*Message, error)`
Creates a new answer message.

#### `NewEmptyMessage() *Message`
Creates a new `Message` object with no AVPs.

#### `SendMessage(peer *node.Node, msg *Message) error`
Serializes the message and sends it to the specified peer.

#### `RecvMessage(peer *node.Node, wait bool) (*Message, error)`
Receives a message from a peer and deserializes it into a `Message` object. `
wait` specifies whether to block until a message is received.

#### `ReplyToMessage(msg *Message) (*Message, error)`
Generates a response message for the given message.

#### `BytesToMessage(data []byte) (*Message, error)`
Directly converts a byte slice into a `Message` object.

---

### Message Operations

#### `AddAvp(avpId any) error`
Adds an AVP to the message. `avpId` can be:
- `*Avp`: An existing AVP instance.
- `string`: The AVP name (e.g., "Session-Id").
- `uint32`/`int`: The AVP code.

#### `RemoveAvp(avpId any) error`
Removes an AVP from the message matching the provided identifier.

#### `GetAvp(avpId any) (*Avp, error)`
Retrieves the first AVP matching the provided identifier.

#### `GetAvpNth(avpId any, index int) (*Avp, error)`
Retrieves the Nth AVP matching the identifier (1-based index).

#### `GetAvpValue(avpId any) (any, error)`
Convenience method to retrieve the decoded value of the first matching AVP.

#### `Response() (*Message, error)`
Generates an answer message for the current request, copying the `Session-Id` and transaction identifiers.

#### `Serialize() ([]byte, error)`
Converts the message to wire format. Results are cached.

#### `Deserialize(data []byte) error`
Parses wire format bytes into the message structure.

#### `IsRequest() bool` / `IsError() bool` / `IsProxyable() bool` / `IsRetransmition() bool`
Check flags to determine message type.

#### `Trace()`
Outputs the message in human-readable format to stdout.

---

### AVP Operations

#### `SetValue(value any) error`
Validates the `value` Go type against the dictionary and encodes it using the appropriate codec.

#### `Value() any`
Returns the decoded Go value of the AVP.

#### `IsVendorSpec() bool`
Returns true if the Vendor-Specific (V) flag is set.

#### `IsMandatory() bool`
Returns true if the Mandatory (M) flag is set.

#### `IsProtected() bool`
Returns true if the Protected (P) flag is set.

#### `IsGrouped() bool`
Returns true if the AVP is of type `Grouped`.

#### `Dump(shift ...int)`
Prints the AVP and its value (including nested AVPs for grouped types) to stdout.

---

## Workflow Example

```go
// 1. Initialize Environment
d, _ := diameter.New(diameter.ModeTransaction)
d.LoadDict("dictionary.pkl", 0)
d.LoadData("avp-data.yaml")
d.LoadPeers("peers.yaml")

// 2. Create a Request
msg, _ := d.NewRequest(16777251, "UL")

// 3. Set AVP Values
avp, _ := msg.GetAvp("Session-Id")
avp.SetValue("my-session-123")

// 4. Create to Peer
peer, _ := d.NewPeer("server", "192.168.0.1", 3868, "sctp", 5)

// 5. Send to Peer
d.SendMessage(peer, msg)

// 6. Receive Response
resp, _ := d.RecvMessage(peer, true)
fmt.Printf("Result Code: %v\n", resp.GetAvpValue("Result-Code"))
```
