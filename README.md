
# TGDP - Trafic Generator for Diameter Protocol

## Description
A Diameter packet generator is a tool designed to create customizable Diameter protocol messages,
enabling network engineers and developers to simulate real-world traffic, validate system behavior, 
and troubleshoot Diameter-based applications and infrastructure efficiently.

## Features
- Creating Diameter messages according to external description in Pkl lang
- Messaging with peers
- REPL interactive mode
- Writing PCAP files
- Simple Diameter server
- Built-in scripting in Lua language
- Support Linux or MacOS

## Current limitations
- SCTP/TCP "header parameters" (TTL, flags, Seq, ...) storing in PCAP are fakes. Only IP addresses and ports are real.
- SCTP transport not supported on MacOS due to lack of support in the underlying libraries.
- SCTP multichuncking not supported.
- No handling for the messages T-flag.
- No E2E encryption suppports.
- No Route-Path or Route-Record AVP handling.
- No proxy handling (Proxy-Info, Route-Record)

## Usage
How to use TGDP - see [User Guide](https://github.com/LonelyCat/tgdp/blob/main/docs/User-Guide.md)

## License
[MIT](https://choosealicense.com/licenses/mit/) - feel free to use this project for any non-commercial purposes

## Contact
Project Link: [https://github.com/LonelyCat/tgdp](https://github.com/LonelyCat/tgdp)

## Support
For support, email alexander.kefeli@gmail.com
