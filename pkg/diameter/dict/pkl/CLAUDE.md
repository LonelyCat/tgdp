# CLAUDE.md - pkg/diameter/dict/pkl

Pkl source files defining the Diameter protocol dictionary data structures.

## Files

| File | Description |
|------|-------------|
| `Diameter.pkl` | Core type definitions: bit flags, data types, and base classes (App, Command, Avp, AvpRule, Enum, Group) |
| `Core.pkl` | Diameter dictionary core structure for Go-lang binding |
| `apps.pkl` | Diameter applications: Common Messages, S6a, Sh, S6c, SLg, SLh, Cx with command definitions |
| `avps.pkl` | Diameter AVP (Attribute-Value Pair) definitions including grouped and enumerated types |
| `dictionary.pkl` | Entry point that aggregates Apps, Avps, CmdFlags, AvpFlags, AvpTypes |

## Core Types (Core.pkl)

```pkl
class CmdBitFlags { R, P, E, T }    // Command flags
class AvpBitFlags { V, M, P }       // AVP flags
class AvpDataTypes { OctetString, Integer32, ... } // 16 data types

class App { id, name, vnd, vnd_id, cmds }
class Command { code, name, short, flags, request, answer }
class Avp { code, name, flags, vnd_id, type, enum?, group? }
class AvpRule { name, required, max? }
class Enum { items: Listing<Item> }
class Group { members: Listing<AvpRule> }
```

## Applications (apps.pkl)

- **Common Messages** (id=0): CE, DW, DP commands
- **S6a** (id=16777251): UL, CL, AI, ID, DD, PU, RE, NO commands
- **Sh** (id=16777217): UD, PU, SN, PN commands
- **S6c** (id=16777312): SR command
- **SLg** (id=16777255): PL, LR commands
- **SLh** (id=16777291): RI command
- **Cx** (id=16777216): UA, SA commands

## Regeneration

Run `make pkl2go` to regenerate Go code from these Pkl definitions.
