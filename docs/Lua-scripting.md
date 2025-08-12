
# Internal API for Lua scripting with TGDP

## Overview
TGDP provides an advanced way to create a scenario using the Lua language.
To execute a Lua script using TGDP, type
```bash
tgdp [flags] @my_script.lua
```

Diameter support for Lua scripting implemented by TGDP as "virtual" Lua module called 'diameter'.

```lua
local dia = require("diameter")
-- To do something ...
--
```

## API Data Types
The API provides the following data types:
* **`peer`** - remote peers handling
* **`message`** - Diameter messages handling
* **`avp`** - Diameter AVPs handling

The module provides a `new(...)` method to create an instance for each data type.
The module also provides a `get(...)` method for `peer` and `message` types to get an existing instance from global definitions.

```lua
local dia = require("diameter")

peer, err = dia.peer.new(...)
msg, err = dia.message.new(...)
```

## Module constants
Message's flags:
* MSG_FLAG_REQUEST
* MSG_FLAG_PROXYABLE
* MSG_FLAG_ERROR
* MSG_FLAG_RETRANSMITION

AVP's flags:
* AVP_FLAG_VENDOR_SPECIFIC
* AVP_FLAG_MANDATORY
* AVP_FLAG_PROTECTED

AVP's data types:
* AVP_TYPE_OCTET_STRING
* AVP_TYPE_INTEGER32
* AVP_TYPE_INTEGER64
* AVP_TYPE_UNSIGNED32
* AVP_TYPE_UNSIGNED64
* AVP_TYPE_FLOAT32
* AVP_TYPE_FLOAT64
* AVP_TYPE_ADDRESS
* AVP_TYPE_TIME
* AVP_TYPE_UTF8_STRING
* AVP_TYPE_IDENTITY
* AVP_TYPE_IP_FILTER_RULE
* AVP_TYPE_QOS_FILTER_RULE
* AVP_TYPE_ENUMERATED
* AVP_TYPE_GROUPED


## Data type mapping Diameter to Lua
|Diameter|Lua  |
|--|--|
| Integer32 | Number|
| Unsigned32 | Number|
| Integer64 | Number|
| Unsigned64 | Number|
| Float32 | Number|
| Float64 | Number|
| UTF8String |String|
| Identity |String|
| URI |String|
| IPFilterRule |String|
| QoSFilterRule |String|
| OctetString |String|
| UTF8String |String|
| Address | String |
| Time | String |
| Enumerated | Number|
| Grouped | - |


## `Peer`
### Module Functions

#### `<module>.peer.new(name, address, port, transport) (peer, error)`
##### Description
Create new peer instance with specified name, address, port and transport protocol.

##### Parameters:
* **`name`** - peer name
* **`address`** - IP address or DNS name
* **`port`** - SCTP or TCP port
* **`transport`** - "SCTP" or "TCP"

##### Return values:
* **`peer`** - peer object if no an error
* **`error`** - an error if occurred

##### Example
```lua
dia = require("diameter")
peer, err = dia.node.new("peer1", "127.0.0.1", 3868, "sctp")
```

#### `<module>.peer.get(name) (peer, error)`
##### Description
Return an existing peer instance with specified name.

##### Parameters:
* **`name`** - peer name

##### Return values:
* **`peer`** - peer object
* **`error`** - an error if occurred

##### Example
```lua
dia = require("diameter")
mme, err = dia.node.get("mme_peer")
```
### Properties

 `name` - peer's name
 `address` - peer's address
 `port` - peer's port
 `protocol` - peer's protocol

  **Note**: all properties are read only

### Methods

#### `connect() error`
##### Description
Connect to remote node.

##### Return values:
* **`error`** - an error if occurred

##### Example
```lua
dia = require("diameter")
hss, err = dia.peer.new("hss", "127.0.0.1", 3868, "sctp")
err = hss:connect()
```

#### `disconnect() error`
##### Description
Disconnects from the node.

##### Return values:
* **`error`** - an error if occurred

##### Example
```lua
err = mme:disonnect()
```

#### `send_to(message) error`
##### Description
Send a message to remote node.

##### Parameters:
* **`message`** - an instance of 'message' data type

##### Return values:
* **`error`** - an error if occurred

##### Example
```lua
err = hss:send_to(ulr) -- ulr - instance of 'message'
```

#### `recv_from() (message, error)`
##### Description
Receive a message from remote node.

##### Return values:
* **`message`** - an instance of 'message' data type if no a error
* **`error`** - an error if occurred

##### Example
```lua
ula, err = hss:recv_from()
```


## `Message`
### Module Functions

#### `<module>.message.new(app, cmd, request) (message, error)`
##### Description
Create new message instance for application, with command code and flags, without an AVP. Use method `add_avp` for adding an one.

##### Parameters:
* **`app`** - application id
* **`cmd`** - command code
* **`request`** - boolean flag defines the message type: request or not

##### Return values:
* **`message`** - message object if no an error
* **`error`** - an error if occurred

##### Example
```lua
dia = requre("diameter)
ulr, err = dia.message.new("S6a", "UL", true) -- Update Location request
```

#### `<module>.message.get(app, cmd, request) (message, error)`
##### Description
Create new message instance for application, with command code and flags, with AVPs, based on Pkl definition.

##### Parameters:
* **`app`** - application id
* **`cmd`** - command code
* **`request`** - boolean flag defines the message type: request or not

##### Return values:
* **`message`** - message object if no an error
* **`error`** - an error if occurred

##### Example
```lua
dia = requre("diameter)
clr, err = dia.message.get("S6a", "CL", false) -- Cancel Location answer
```

### Properties

 `app_id` - message application ID
 `app_name` - message application name (read-only)
 `cmd_code` - message command code
 `flags` - message flags
 `hop_by_hop` - message "hop by hop" value
 `end_to_end` - message "end to end" value
 `avps` - AVP message list

### Methods

#### `add_avp(avp) error`
##### Description
Add an AVP to message.

##### Parameters:
* **`avp`** - AVP's string name, or numeric id, or an existing instance of 'avp' data type<br>
If **`avp`** defined by name or id it will be automatically created based on Pkl definitions.

##### Return values:
* **`error`** - an error if occurred.

##### Example
```lua
ulr, err = dia.message.new("S6a", "UL", true)
ulr:add_avp("ULR-FLags") -- add AVP by name
ulr:add_avp(1) -- add AVP by id (User-Name = 1)
ulr:add_avp(avp_ulr_flags) -- add AVP by an object (created before)
```

#### `get_avp(avp) error`
##### Description
Return an AVP instance from message.

##### Parameters:
* **`avp`** - AVP's string name, or numeric id.

##### Return values:
* **`avp`** - an instance of 'avp' datatype.
* **`error`** - an error if occurred.

##### Example
```lua
ulr, err = dia.message.get("S6a", "UL", true)
imsi, err = ulr:get_avp("User-Name")
imsi, err = ulr:get_avp(1)
```

#### `remove_avp(avp) error`
##### Description
Remove an AVP instance from message.

##### Parameters:
* **`avp`** - AVP's string name, or numeric id, or an existing instance of 'avp' data type<br>

##### Return values:
* **`error`** - an error if occurred.

##### Example
```lua
ulr, err = dia.message.get("S6a", "UL", true)
err = ulr.remove_avp("Route-Record")
```

#### `get_avp_value(avp) <AVP value>, error`
##### Description
Get AVP value from message.

##### Parameters:
* **`avp`** - AVP's string name, or numeric id, or an existing instance of 'avp' data type<br>

##### Return values:
* **`AVP value`** - if no an error.
* **`error`** - if an error.

##### Example
```lua
ula, err = peer:recv_from()
result_code, err = ula:get_avp_value("Result-Code")
```

#### `set_avp_value(avp, value) error`
##### Description
Set AVP value

##### Parameters:
* **`avp`** - AVP's string name, or numeric id, or an existing instance of 'avp' data type<br>
* **`value`** - Lua data for AVP assignment

##### Return values:
* **`error`** - if an error.

##### Example
```lua
ulr, err = dia.message.get("S6a", "UL", true)
err = ulr:set_avp_value("Route-Record", "dra01.mnc040.mcc250.3gppnetwork.org")
```

#### `is_request() bool`
##### Description
Check if message is request.

##### Return values:
* **`bool`** - true if message is request, false otherwise.

##### Example
```lua
is_req_msg = msg:is_request()
```


## `AVP`
### Module Functions

#### `<module>.avp.new(name, code, flags, vendor_id, type) (avp, error)`
##### Description
Create new AVP instance

##### Parameters:
* **`name`** - AVP name
* **`cmd`** - AVP code
* **`flags`** - AVP flags (see `Constants` for more details)
* **`vendor_id`** - vendor id
* **`type`** - AVP data type (see `Constants` for more details)

##### Return values:
* **`avp`** - app object if no an error
* **`error`** - an error if occurred

##### Example
```lua
dia = requre("diameter)
imsi, err = dia.avp.new("User-Name", 1, dia.AVP_FLAG_MANDATORY, 0, dia.AVP_TYPE_UTF8STRING)
```

#### `<module>.avp.get(avp) (avp, error)`
##### Description
Create new AVP instance based on Pkl definition.

##### Parameters:
* **`id`** - AVP id: string name or numerical id.

##### Return values:
* **`avp`** - message object if no an error
* **`error`** - an error if occurred

##### Example
```lua
dia = requre("diameter)
msisdn, err = dia.avp.get("MSISDN")
imsi, err = dia:avp.get(1) -- 1 - User-Name
```

### Properties

 `code` - AVP code
 `flags` - AVP flags
 `name` - AVP name
 `vendor_id` - AVP vendor id
 `members` - grouped AVP members

### Methods

#### `get_value(avp) (<AVP value>, error)`
##### Description
Get AVP value.

##### Parameters:
* **`avp`** - AVP's string name, or numeric id, or an existing instance of 'avp' data type<br>

##### Return values:
* **`AVP value`** - if no an error.
* **`error`** - if an error.

##### Example
```lua
ula, err = peer:recv_from()
result_code_avp, err = ula:get_avp("Result-Code")
result_code, err = result_code_avp:get_value()
```

#### `set_value(value) error`
##### Description
Set AVP value

##### Parameters:
* **`value`** - Lua data for AVP assignment

##### Return values:
* **`error`** - if an error.

##### Example
```lua
ula, err = peer:recv_from()
result_code_avp, err = ula:get_avp("Result-Code")
result_code_avp:set_value(2001)
```

#### `is_grouped() bool`
##### Description
Check if AVP is grouped type.

##### Return values:
* **`bool`** - true if AVP is grouped, false otherwise.

##### Example
```lua
is_grouped_avp = avp:is_grouped()
```

#### `is_mandatory() bool`
##### Description
Check if AVP is mandatory.

##### Return values:
* **`bool`** - true if AVP is mandatory, false otherwise.

##### Example
```lua
is_mandatory_avp = avp:is_mandatory()
```
