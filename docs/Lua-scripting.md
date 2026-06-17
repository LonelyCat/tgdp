# Internal API for Lua scripting with TGDP

## Table of contents

- [Overview](#overview)
- [API Data Types](#api-data-types)
- [Module Constants](#module-constants)
  - [Message Flags](#message-flags)
  - [AVP Flags](#avp-flags)
  - [AVP Data Types](#avp-data-types)
- [Diameter to Lua Data Type Mapping](#diameter-to-lua-data-type-mapping)
- [`Peer`](#peer)
  - [Module level functions](#module-level-functions)
    - [`dia.peer.new(name, address, port, transport) -> (peer, err)`](#diapeernewname-address-port-transport-peer-err)
    - [`dia.peer.fetch(name) -> (peer, err)`](#diapeerfetchname-peer-err)
  - [Properties](#properties)
  - [Methods](#methods)
    - [`peer:connect() -> err`](#peerconnect-err)
    - [`peer:disconnect() -> err`](#peerdisconnect-err)
    - [`peer:send_to(message) -> err`](#peersendtomessage-err)
    - [`peer:recv_from() -> (message, err)`](#peerrecvfrom-message-err)
- [`Message`](#message)
  - [Module level functions](#module-level-functions-2)
    - [`dia.message.new(app, cmd, is_request) -> (message, err)`](#diamessagenewapp-cmd-isrequest-message-err)
    - [`dia.message.fetch(app, cmd, is_request) -> (message, err)`](#diamessagefetchapp-cmd-isrequest-message-err)
  - [Properties](#properties-2)
  - [Methods](#methods-2)
    - [`message:add_avp(avp) -> err`](#messageaddavpavp-err)
    - [`message:get_avp(avp_id) -> (avp, err)`](#messagegetavpavpid-avp-err)
    - [`message:remove_avp(avp_id) -> err`](#messageremoveavpavpid-err)
    - [`message:get_avp_value(avp_id) -> (value, err)`](#messagegetavpvalueavpid-value-err)
    - [`message:set_avp_value(avp_id, value) -> err`](#messagesetavpvalueavpid-value-err)
    - [`message:is_request() -> boolean`](#messageisrequest-boolean)
- [`AVP`](#avp)
  - [Module level functions](#module-level-functions-3)
    - [`dia.avp.new(name, code, flags, vendor_id, type) -> (avp, err)`](#diaavpnewname-code-flags-vendorid-type-avp-err)
    - [`dia.avp.fetch(avp_id) -> (avp, err)`](#diaavpfetchavpid-avp-err)
  - [Properties](#properties-3)
  - [Methods](#methods-3)
    - [`avp:get_value() -> (value, err)`](#avpgetvalue-value-err)
    - [`avp:set_value(value) -> err`](#avpsetvaluevalue-err)
    - [`avp:is_grouped() -> boolean`](#avpisgrouped-boolean)
    - [`avp:is_mandatory() -> boolean`](#avpismandatory-boolean)

## Overview
TGDP is a command-line tool for testing Diameter protocol implementations.
It provides an advanced way to create scenarios using the Lua language.
To execute a Lua script using TGDP, type:
```bash
tgdp [flags] @my_script.lua [parameters]
```

Diameter support for Lua scripting is implemented by TGDP as a "virtual" Lua module called `diameter`.

```lua
local dia = require("diameter")
-- Ready to use the API
```

## API Data Types
The API provides the following data types:
* **`peer`** - Handles connections to remote Diameter peers.
* **`message`** - Handles creation and manipulation of Diameter messages.
* **`avp`** - Handles creation and manipulation of Diameter Attribute-Value-Pairs (AVPs).

The module provides a `new(...)` method to create a new instance for each data type.
The module also provides a `get(...)` method for `peer` and `message` types to get a pre-configured instance from global definitions, writing in Pkl.

```lua
local dia = require("diameter")

-- Create a new peer from scratch
local peer, err = dia.peer.new(...)

-- Create a new message from scratch
local msg, err = dia.message.new(...)
```

## Module Constants
All constants are available directly from the `diameter` module (e.g., `dia.MSG_FLAG_REQUEST`).

### Message Flags
* `MSG_FLAG_REQUEST`
* `MSG_FLAG_PROXYABLE`
* `MSG_FLAG_ERROR`
* `MSG_FLAG_RETRANSMISSION`

### AVP Flags
* `AVP_FLAG_VENDOR_SPECIFIC`
* `AVP_FLAG_MANDATORY`
* `AVP_FLAG_PROTECTED`

### AVP Data Types
* `AVP_TYPE_OCTET_STRING`
* `AVP_TYPE_INTEGER32`
* `AVP_TYPE_INTEGER64`
* `AVP_TYPE_UNSIGNED32`
* `AVP_TYPE_UNSIGNED64`
* `AVP_TYPE_FLOAT32`
* `AVP_TYPE_FLOAT64`
* `AVP_TYPE_ADDRESS`
* `AVP_TYPE_TIME`
* `AVP_TYPE_UTF8_STRING`
* `AVP_TYPE_IDENTITY`
* `AVP_TYPE_IP_FILTER_RULE`
* `AVP_TYPE_QOS_FILTER_RULE`
* `AVP_TYPE_ENUMERATED`
* `AVP_TYPE_GROUPED`


## Diameter to Lua Data Type Mapping
| Diameter Data Type | Lua Data Type |
|--------------------|---------------|
| Integer32 | `number` |
| Unsigned32 | `number` |
| Integer64 | `number` |
| Unsigned64 | `number` |
| Float32 | `number` |
| Float64 | `number` |
| OctetString | `string` |
| UTF8String | `string` |
| Identity | `string` |
| URI | `string` |
| IPFilterRule | `string` |
| QoSFilterRule | `string` |
| Address | `string` |
| Time | `string` |
| Enumerated | `number` |
| Grouped | `AVP object` |


## `Peer`
A `peer` object represents a remote Diameter node.

### Module level functions

#### `dia.peer.new(name, address, port, transport) -> (peer, err)`
##### Description
Creates a new peer instance with a specified name, address, port, and transport protocol.

##### Parameters:
* `name` (`string`): The peer name.
* `address` (`string`): The IP address or DNS name.
* `port` (`number`): The SCTP or TCP port.
* `transport` (`string`): `"SCTP"` or `"TCP"`.

##### Return values:
* `peer`: The new peer object if successful.
* `err`: An error string if an error occurred.

##### Example
```lua
local dia = require("diameter")
local peer, err = dia.peer.new("peer1", "127.0.0.1", 3868, "sctp")
if err then
-- handle error
end
```

#### `dia.peer.fetch(name) -> (peer, err)`
##### Description
Returns a pre-configured peer instance with the specified name (defined globally, e.g., in a config file).

##### Parameters:
* `name` (`string`): The name of the peer to retrieve.

##### Return values:
* `peer`: The peer object if found.
* `err`: An error string if an error occurred.

##### Example
```lua
local dia = require("diameter")
local mme, err = dia.peer.fetch("mme_peer")
if err then
-- handle error
end
```
### Properties
* `name` (`string`): The peer name.
* `address` (`string`): The peer address.
* `port` (`number`): The peer port.
* `transport` (`string`): The peer transport protocol (`"SCTP"` or `"TCP"`).

**Note**: All properties are read-only.

### Methods

#### `peer:connect() -> err`
##### Description
Establishes a connection to the remote peer.

##### Return values:
* `err`: An error string if an error occurred.

##### Example
```lua
local dia = require("diameter")
local hss, err = dia.peer.new("hss", "127.0.0.1", 3868, "sctp")
err = hss:connect()
if err then
-- handle error
end
```

#### `peer:disconnect() -> err`
##### Description
Disconnects from the remote peer.

##### Return values:
* `err`: An error string if an error occurred.

##### Example
```lua
err = mme:disconnect()
if err then
-- handle error
end
```

#### `peer:send_to(message) -> err`
##### Description
Sends a message to the remote peer.

##### Parameters:
* `message` (`message`): An instance of a `message` object.

##### Return values:
* `err`: An error string if an error occurred.

##### Example
```lua
-- ulr is an instance of a 'message'
local err = hss:send_to(ulr)
if err then
-- handle error
end
```

#### `peer:recv_from() -> (message, err)`
##### Description
Receives a message from the remote peer. This is a blocking call.

##### Return values:
* `message`: A `message` object if successful.
* `err`: An error string if an error occurred.

##### Example
```lua
local ula, err = hss:recv_from()
if err then
-- handle error
end
```


## `Message`
A `message` object represents a Diameter message.

### Module level functions

#### `dia.message.new(app, cmd, is_request) -> (message, err)`
##### Description
Creates a new, empty message instance for a given application, with a command code and flags.
Use the `message:add_avp()` method to add AVPs.

##### Parameters:
* `app` (`string` or `number`): The application name or ID.
* `cmd` (`string` or `number`): The command name or code.
* `is_request` (`boolean`): `true` if the message is a request, `false` for an answer.

##### Return values:
* `message`: The new message object if successful.
* `err`: An error string if an error occurred.

##### Example
```lua
local dia = require("diameter")
-- Create an Update-Location-Request
local ulr, err = dia.message.new("S6a", "UL", true)
```

#### `dia.message.fetch(app, cmd, is_request) -> (message, err)`
##### Description
Creates a new message instance pre-populated with AVPs based on the Pkl dictionary definition.

##### Parameters:
* `app` (`string` or `number`): The application name or ID.
* `cmd` (`string` or `number`): The command name or code.
* `is_request` (`boolean`): `true` if the message is a request, `false` for an answer.

##### Return values:
* `message`: The new message object if successful.
* `err`: An error string if an error occurred.

##### Example
```lua
local dia = require("diameter")
-- Create a Cancel-Location-Answer, pre-filled with required AVPs
local clr, err = dia.message.fetch("S6a", "CL", false)
```

### Properties
* `app_id` (`number`): The message Application-ID.
* `app_name` (`string`): The message application name (read-only).
* `cmd_code` (`number`): The message command code.
* `flags` (`number`): The message flags.
* `hop_by_hop` (`number`): The message Hop-by-Hop Identifier.
* `end_to_end` (`number`): The message End-to-End Identifier.
* `avps` (`table`): A list of AVP objects in the message.

### Methods

#### `message:add_avp(avp) -> err`
##### Description
Adds an AVP to the message.

##### Parameters:
* `avp` (`string`, `number`, or `avp`): The AVP to add. This can be:
* An AVP name (e.g., `"User-Name"`).
* An AVP code (e.g., `1`).
* An existing `avp` object.

If a name or code is provided, the AVP is created based on the Pkl dictionary definition.

##### Return values:
* `err`: An error string if an error occurred.

##### Example
```lua
local ulr, err = dia.message.new("S6a", "UL", true)
-- Add AVP by name
ulr:add_avp("ULR-Flags")
-- Add AVP by code (User-Name = 1)
ulr:add_avp(1)
-- Add a pre-created AVP object
local my_avp, _ = dia.avp.get("My-AVP")
ulr:add_avp(my_avp)
```

#### `message:get_avp(avp_id) -> (avp, err)`
##### Description
Retrieves the first matching AVP instance from the message.

##### Parameters:
* `avp_id` (`string` or `number`): The AVP name or code.

##### Return values:
* `avp`: An `avp` object if found.
* `err`: An error string if an error occurred.

##### Example
```lua
ulr, _ = dia.message.get("S6a", "UL", true)
imsi, err = ulr:get_avp("User-Name")
imsi_by_code, err = ulr:get_avp(1)
```

#### `message:remove_avp(avp_id) -> err`
##### Description
Removes the first matching AVP instance from the message.

##### Parameters:
* `avp_id` (`string` or `number`): The AVP name or code.

##### Return values:
* `err`: An error string if an error occurred.

##### Example
```lua
local ulr, _ = dia.message.get("S6a", "UL", true)
local err = ulr:remove_avp("Route-Record")
```

#### `message:get_avp_value(avp_id) -> (value, err)`
##### Description
A convenient shortcut to get an AVP value directly from a message.

##### Parameters:
* `avp_id` (`string` or `number`): The AVP name or code.

##### Return values:
* `value`: The AVP value if found. The Lua type depends on the AVP data type.
* `err`: An error string if an error occurred.

##### Example
```lua
local ula, _ = peer:recv_from()
local result_code, err = ula:get_avp_value("Result-Code")
```

#### `message:set_avp_value(avp_id, value) -> err`
##### Description
A convenient shortcut to set an AVP value directly in a message.

##### Parameters:
* `avp_id` (`string` or `number`): The AVP name or code.
* `value`: The value to set. The Lua type should be compatible with the AVP data type.

##### Return values:
* `err`: An error string if an error occurred.

##### Example
```lua
local ulr, _ = dia.message.fetch("S6a", "UL", true)
err = ulr:set_avp_value("User-Name", "1234567890")
```

#### `message:is_request() -> boolean`
##### Description
Checks if the message is a request.

##### Return values:
* `boolean`: `true` if the message is a request, `false` otherwise.

##### Example
```lua
local is_req_msg = msg:is_request()
```


## `AVP`
An `avp` object represents a Diameter Attribute-Value-Pair.

### Module level functions

#### `dia.avp.new(name, code, flags, vendor_id, type) -> (avp, err)`
##### Description
Creates a new AVP instance from scratch.

##### Parameters:
* `name` (`string`): The AVP name.
* `code` (`number`): The AVP code.
* `flags` (`number`): The AVP flags (e.g., `dia.AVP_FLAG_MANDATORY`).
* `vendor_id` (`number`): The Vendor-ID. Use `0` if not vendor-specific.
* `type` (`number`): The AVP data type (e.g., `dia.AVP_TYPE_UTF8_STRING`).

##### Return values:
* `avp`: The new AVP object if successful.
* `err`: An error string if an error occurred.

##### Example
```lua
local dia = require("diameter")
local imsi, err = dia.avp.new("User-Name", 1, dia.AVP_FLAG_MANDATORY, 0, dia.AVP_TYPE_UTF8_STRING)
```

#### `dia.avp.fetch(avp_id) -> (avp, err)`
##### Description
Creates a new AVP instance with value based on its Pkl dictionary definition.

##### Parameters:
* `avp_id` (`string` or `number`): The AVP name or code.

##### Return values:
* `avp`: The new AVP object if successful.
* `err`: An error string if an error occurred.

##### Example
```lua
dia = require("diameter")
msisdn, err = dia.avp.fetch("MSISDN")
imsi, err = dia.avp.fetch(1) -- 1 is the code for "User-Name" AVP
```

### Properties
* `code` (`number`): The AVP code.
* `flags` (`number`): The AVP flags.
* `name` (`string`): The AVP name.
* `vendor_id` (`number`): The AVP Vendor-ID.
* `members` (`table`): For grouped AVPs, a list of member AVP objects.

### Methods

#### `avp:get_value() -> (value, err)`
##### Description
Gets the AVP value.

##### Return values:
* `value`: The AVP value. The Lua type depends on the AVP data type.
* `err`: An error string if an error occurred.

##### Example
```lua
ula, _ = peer:recv_from()
result_code_avp, _ = ula:get_avp("Result-Code")
result_code, err = result_code_avp:get_value()
```

#### `avp:set_value(value) -> err`
##### Description
Sets the AVP value.

##### Parameters:
* `value`: The value to set. The Lua type should be compatible with the AVP data type.

For grouped AVPs:
* `value` should be a table of AVP objects.
* `value` adds the member AVPs to the grouped AVP, not replacing any existing member AVPs.

##### Return values:
* `err`: An error string if an error occurred.

##### Example
```lua
ula, _ = dia.message.get("S6a", "UL", false)
result_code_avp, _ = ula:get_avp("Result-Code")
result_code_avp:set_value(2001)
```

#### `avp:is_grouped() -> boolean`
##### Description
Checks if the AVP is of type Grouped.

##### Return values:
* `boolean`: `true` if the AVP is grouped, `false` otherwise.

##### Example
```lua
is_grouped_avp = avp:is_grouped()
```

#### `avp:is_mandatory() -> boolean`
##### Description
Checks if the AVP mandatory ('M') flag is set.

##### Return values:
* `boolean`: `true` if the AVP is mandatory, `false` otherwise.

##### Example
```lua
is_mandatory_avp = avp:is_mandatory()
```
