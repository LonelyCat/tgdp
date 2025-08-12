# Diameter Definition in Pkl

**Note**: Knowledge of Apple's Pkl configuration description language is required.
Visit `www.pkl-lang.org` for more details.

The Diameter protocol definition consists of following files:
* `Diameter.pkl` – definition of Diameter data structures
* `dictionary.pkl` – Diameter dictionary loading by TGDP
* `apps.pkl` – Diameter applications definition
* `avps.pkl` – Diameter AVPs definition

The `Diameter.pkl` and `dictionary.pkl` files should not be modified by the user.
The user can change the composition of generated messages by editing the `apps.pkl` and `avps.pkl` files.

**Note**: check  `apps.pkl` and `avps.pkl` after editing. Otherwise TGDP may fail to start.

## Applications definitions - apps.pkl
Format for definition an Application in BNF format:
```
App ::= id name vnd vnd_id cmds
  id ::= <number>
  name ::= <string>
  vnd ::= <string>
  vnd_id ::= <number>
  cmds ::= Command [Command]

Command ::= code name short flags request answer
  code ::= <number>
  name ::= <string>
  short ::= <string>
  flags ::= <number>
  request ::= AvpRule [AvpRule]
  answer ::= AvpRule [AvpRule]

AvpRule ::= name required max
  name ::= <string>
  required ::= <true | false>
  max ::= <number>
```
Application parameters:
* `id` – 3GPP identifier
* `name` – 3GPP name (recommended, but may differ)
* `vnd` – vendor name
* `vnd_id` – vendor identifier
* `cmds` – list of possible commands

Command parameters:
* `code` – code
* `name` – full name
* `short` – abbreviation
* `flags` – bit flags
* `request` – list of AVPs for the request
* `answer` – list of AVPs for the answer

AVP parameters:
* `name` – AVP name
* `required` – flag indicating mandatory presence in the message
* `max` – maximum allowed number of occurrences (if 0 or absent - unlimited)

```pkl
new App {
  id = 16777251
  name = "S6a"
  vnd = "3GPP"
  vnd_id = 10415
  cmds = new Listing {
  // Update Location Request/Answer
  new Command { code=316 name="Update Location" short="UL" flags=P
    request = new Listing {
      new Avp { name="Session-Id" required=true max=1 }
      new Avp { name="DRMP" required=false max=1 }
      new Avp { name="Vendor-Specific-Application-Id" required=true max=1 }
      ...
    }

  answer = new Listing {
    new Avp { name="Session-Id" required=true max=1 }
    new Avp { name="Result-Code" required=false max=1 }
    new Avp { name="Experimental-Result" required=false max=1 }
    new Avp { name="Error-Diagnostic" required=false max=1 }
    new Avp { name="Auth-Session-State" required=true max=1 }
    ...
    }
  }
}
```

## AVPs definitions - avps.pkl
Format for definition an AVP in BNF format:

```
AVP ::= code name flags [vnd_id] type [enum] | [group]
  code ::= <number>
  name ::= <string>
  vnd_id ::= <number>
  type ::= OctetString | Integer32 | Integer64 | Unsigned32 | Unsigned64 | Float32 | Float64 | Address | Time | UTF8String | Identity | URI | Enumerated | IPFilterMember | QoSFilterMember | Grouped

enum ::= Enum
Enum :: = Items
Items ::== Item [Item]
Item ::= code name
  code ::= <number>
  name ::= <string>

group ::= Group
Group ::= Members
Members ::= AvpRule [AvpRule]
AvpRule ::= name required max
  name ::= <string>
  required ::= <true | false>
  max ::= <number>
```

AVP parameters:
* `code` – 3GPP code
* `name` – 3GPP full name
* `flags` – bit flags
* `vnd_id` – vendor identifier (optional)
* `type` – data type
* `Enum` – data for the Enumerated type
* `Group` – grouped AVP composition for the Grouped type

Item parameters:
* `code` – item code
* `name` – mnemonic name of the item

Member parameters:
* `name` – AVP name
* `required` – mandatory presence flag
* `max` – maximum allowed number of occurrences (if 0 or absent - unlimited)

```pkl
// Simple type AVPs
new Avp { code=1406 name="ULA-Flags" flags=M+V vnd_id=10415 type=Unsigned32 }
new Avp { code=1 name="User-Name" flags=M type=UTF8String }
```

```pkl
// Grouped AVP
new Avp { code=260 name="Vendor-Specific-Application-Id" flags=M type=Grouped
  group = new Group {
    members = new Listing {
      new Member { name="Vendor-Id" required=true max=1 }
      new Member { name="Auth-Application-Id" required=false max=1 }
      new Member { name="Acct-Application-Id" required=false max=1 }
    }
  }
}
```

```pkl
// Enumerated AVP
new Avp { code=1420 name="Cancellation-Type" flags=M+V vnd_id=10415 type=Enumerated
  enum = new Enum {
    items = new Listing {
      new Item { code=0 name="MME_UPDATE_PROCEDURE" }
      new Item { code=1 name="SGSN_UPDATE_PROCEDURE" }
      new Item { code=2 name="SUBSCRIPTION_WITHDRAWWAL" }
      new Item { code=3 name="UPDATE_PROCEDURE" }
      new Item { code=4 name="INITITAL_ATTACH_PROCEDURE" }
    }
  }
}
```
