//
// Project: TGDP - Traffic Generator for Diameter Protocol
// Description: Simple tool for testing and debugging the Diameter protocol
//
// Author: Alexander Kefeli <alexander.kefeli@gmail.com>
//
// File: message.go
// Description: Diameter pkg: message handling
//

package diameter

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log/slog"
	"slices"
	"strings"

	"tgdp/pkg/diameter/diwe"
)

// Consts
//
// MinMessageLen is the minimum length of a Diameter message header (20 bytes).
// This is the size of: Version(1) + Length(3) + Flags(1) + CmdCode(3) + AppId(4) + HopByHop(4) + EndToEnd(4)
const MinMessageLen = uint32(20)

// Types
//
// Message represents a Diameter protocol message.
// A Diameter message consists of a fixed-size header followed by zero or more AVPs.
// The header contains version, length, flags, command code, application ID,
// and hop-by-hop and end-to-end identifiers for transaction tracking.
type Message struct {
	// Public fields (serialized to wire format)

	// Version is the Diameter protocol version (currently 1).
	Version uint8
	// Length is the total message length in bytes (header + all AVPs).
	// This is a 24-bit field (bits 0-23 of the length field).
	Length uint32
	// AppId is the Diameter Application ID that defines the message type.
	AppId uint32
	// Flags contains the message flags (R, P, E, T bits).
	Flags uint8
	// CmdCode is the Command Code identifying the specific message type.
	CmdCode uint32
	// HopByHop is a unique identifier for this transaction, matching request and answer.
	HopByHop uint32
	// EndToEnd is a unique identifier for end-to-end message correlation.
	EndToEnd uint32

	// Hidden fields (not serialized)

	// env is a reference to the Diameter environment for dictionary and codec access.
	env *Diameter
	// avps is the slice of AVPs contained in this message.
	avps []*Avp
	// bytes holds the serialized wire format (cached after first Serialize call).
	bytes []byte
}

// Methods
//
// AddAvp adds an AVP to the message.
// The avpId parameter can be:
//   - *Avp: an existing AVP instance to add
//   - string: AVP name to look up in the dictionary
//   - uint32/int: AVP code to look up in the dictionary
//
// Returns an error if the AVP cannot be found in the dictionary.
func (m *Message) AddAvp(avpId any) error {
	var avp *Avp
	switch avpId := avpId.(type) {
	case *Avp:
		// Already an AVP instance - use directly
		avp = avpId
	default:
		// Look up AVP by name or code from dictionary
		a, err := m.env.GetAvp(avpId)
		if err != nil {
			return err
		}
		avp = a
	}

	// FIXME: check for AVP max count
	m.avps = append(m.avps, avp)
	m.bytes = nil // Invalidate cached serialization

	return nil
}

// RemoveAvp removes an AVP from the message by its identifier.
// The avpId parameter can be an AVP code (uint32/int), name (string), or *Avp pointer.
// Returns ErrMissingAvp if the AVP is not found in the message.
func (m *Message) RemoveAvp(avpId any) error {
	for i, avp := range m.avps {
		if matchAvp(avp, avpId) {
			m.avps = slices.Delete(m.avps, i, i+1)
			m.bytes = nil // Invalidate cached serialization
			return nil
		}
	}

	return &diwe.ErrMissingAvp{Avp: avpId}
}

// Avps retrieves the list of message's AVPs.
// Returns an array of pointers to the message's AVPs.
func (m *Message) Avps() []*Avp {
	return m.avps
}

// GetAvp retrieves the first AVP matching the given identifier.
// The avpId parameter can be an AVP code (uint32/int), name (string), or *Avp pointer.
// Returns ErrMissingAvp if no matching AVP is found.
func (m *Message) GetAvp(avpId any) (*Avp, error) {
	for _, avp := range m.avps {
		if matchAvp(avp, avpId) {
			return avp, nil
		}
	}

	return nil, &diwe.ErrMissingAvp{Avp: avpId}
}

// GetAvpNth retrieves the Nth AVP matching the given identifier.
// The index parameter is 1-based (1 returns first match, 2 returns second, etc.).
// Useful for messages with multiple AVPs of the same code.
// Returns ErrMissingAvp if fewer than 'index' AVPs match.
func (m *Message) GetAvpNth(avpId any, index int) (*Avp, error) {
	n := 1
	for _, avp := range m.avps {
		if matchAvp(avp, avpId) {
			if n == index {
				return avp, nil
			}
			n++
		}
	}

	return nil, &diwe.ErrMissingAvp{Avp: avpId}
}

// GetAvpValue retrieves the value of the first AVP matching the given identifier.
// This is a convenience method that combines GetAvp and Value.
// Returns the AVP value as any, or an error if the AVP is not found.
func (m *Message) GetAvpValue(avpId any) (any, error) {
	avp, err := m.GetAvp(avpId)
	if err != nil {
		return nil, err
	}

	return avp.Value(), nil
}

// Response creates an response/answer message for this request message.
// It uses the same Application ID, Command Code, HopByHop, and EndToEnd identifiers.
// The R (Request) flag is cleared in the reply.
// If the original message contains a Session-Id AVP, it is copied to the reply.
// Returns the new answer message or an error if creation fails.
func (m *Message) Response() (*Message, error) {
	// Create new message as answer (request=false)
	r, err := m.env.NewMessage(m.AppId, m.CmdCode, false, true)
	if err != nil {
		return nil, err
	}

	// Copy Session-Id from request to answer if present
	avpSessionId, err := m.GetAvp("Session-Id")
	if err == nil {
		value := avpSessionId.Value().(string)

		// Reuse the outer 'err' and 'avpSessionId' variables using '=' instead of ':='
		avpSessionId, err = r.GetAvp("Session-Id")
		if err == nil {
			err = avpSessionId.SetValue(value)
			if err != nil {
				return nil, err
			}
		}
	}

	// Copy transaction identifiers from request
	r.HopByHop = m.HopByHop
	r.EndToEnd = m.EndToEnd
	// Clear the Request flag to convert request to answer
	r.Flags = m.Flags & ^m.env.Dict().CmdFlag().R

	// Explicitly return nil for error to avoid returning a stale 'err' state
	return r, nil
}

// Serialize converts the message to its wire format bytes.
// The message header format is:
//   - Byte 0: Version (1 byte)
//   - Bytes 1-3: Length (24-bit unsigned, bits 8-31 of first 4 bytes)
//   - Byte 4: Flags (1 byte)
//   - Bytes 5-7: Command Code (24-bit unsigned)
//   - Bytes 8-11: Application ID (4 bytes)
//   - Bytes 12-15: Hop-by-Hop Identifier (4 bytes)
//   - Bytes 16-19: End-to-End Identifier (4 bytes)
//
// Returns the serialized bytes and any error that occurred.
// Results are cached; subsequent calls return the cached bytes.
func (m *Message) Serialize() ([]byte, error) {
	// Return cached serialization if available
	if m.bytes != nil {
		return m.bytes, nil
	}

	// Pre-allocate buffer with expected length
	buf := bytes.NewBuffer(make([]byte, 0, m.Length))

	// Build the 20-byte header
	var header [MinMessageLen]byte
	// Bytes 0-3: Version (bits 24-31) + Length (bits 0-23)
	binary.BigEndian.PutUint32(header[0:4], (uint32(m.Version)<<24)|(m.Length&0x00FFFFFF))
	// Bytes 4-7: Flags (bits 24-31) + Command Code (bits 0-23)
	binary.BigEndian.PutUint32(header[4:8], (uint32(m.Flags)<<24)|(m.CmdCode&0x00FFFFFF))
	// Bytes 8-11: Application ID
	binary.BigEndian.PutUint32(header[8:12], m.AppId)
	// Bytes 12-15: Hop-by-Hop
	binary.BigEndian.PutUint32(header[12:16], m.HopByHop)
	// Bytes 16-19: End-to-End
	binary.BigEndian.PutUint32(header[16:20], m.EndToEnd)
	buf.Write(header[:])

	// Serialize all AVPs
	for _, avp := range m.avps {
		if err := avp.Serialize(buf); err != nil {
			return nil, err
		}
	}

	// Update length with actual serialized size
	m.Length = uint32(buf.Len())
	// Re-write header with correct length
	binary.BigEndian.PutUint32(buf.Bytes(), (uint32(m.Version)<<24)|(m.Length&0x00FFFFFF))
	m.bytes = buf.Bytes()
	return m.bytes, nil
}

// reset clears the message state for reuse in a pool.
func (m *Message) reset() {
	m.Version = 0
	m.Length = 0
	m.AppId = 0
	m.Flags = 0
	m.CmdCode = 0
	m.HopByHop = 0
	m.EndToEnd = 0
	m.env = nil
	m.avps = m.avps[:0]
	m.bytes = nil
}

// Deserialize parses a Diameter message from wire format bytes.
// It validates the minimum message length, parses the header fields,
// then iteratively deserializes all AVPs from the message body.
//
// The data should be a complete Diameter message including header and all AVPs.
// Returns an error if parsing fails.
func (m *Message) Deserialize(data []byte) error {
	// Validate minimum length
	if len(data) < int(MinMessageLen) {
		return &diwe.ErrMsgTooShort{Len: len(data)}
	}

	var err error
	m.Version, m.Length, m.AppId, m.CmdCode, m.Flags, m.HopByHop, m.EndToEnd, err = m.env.MessageHeader(data)
	if err != nil {
		return err
	}

	// Parse AVPs from message body
	offset := uint32(MinMessageLen)
	for offset < uint32(len(data)) {
		avp := getAvp()
		avp.env = m.env
		o, err := avp.Deserialize(data[offset:])
		if err != nil {
			putAvp(avp)
			return err
		}
		m.avps = append(m.avps, avp)
		offset += o
	}

	// Cache the original wire format
	m.bytes = make([]byte, len(data))
	copy(m.bytes, data)

	return nil
}

// IsCommon returns true if the message is a common message (AppId == 0).
func (m *Message) IsCommon() bool {
	return (m.AppId == 0)
}

// IsRequest returns true if the R (Request) flag is set.
// When true, this message is a request; otherwise it's an answer.
func (m *Message) IsRequest() bool {
	return (m.Flags&m.env.Dict().CmdFlag().R != 0)
}

// IsProxyable returns true if the P (Proxyable) flag is set.
// When true, this message may be proxied or forwarded by intermediaries.
func (m *Message) IsProxyable() bool {
	return (m.Flags&m.env.Dict().CmdFlag().P != 0)
}

// IsError returns true if the E (Error) flag is set.
// When true, this message indicates an error condition.
func (m *Message) IsError() bool {
	return (m.Flags&m.env.Dict().CmdFlag().E != 0)
}

// IsRetransmition returns true if the T (Retransmission) flag is set.
// When true, this message is a retransmission of a previously sent message.
func (m *Message) IsRetransmition() bool {
	return (m.Flags&m.env.Dict().CmdFlag().T != 0)
}

// Bytes returns the cached wire format bytes.
// Returns nil if the message has not been serialized.
func (m *Message) Bytes() []byte {
	return m.bytes
}

// Len returns the message length in bytes.
func (m *Message) Len() uint32 {
	return m.Length
}

// Trace outputs the message to stdout in human-readable format.
// Displays the header fields (version, length, app ID, flags, command code,
// hop-by-hop, end-to-end) and lists all AVPs with their values.
func (m *Message) Trace(shift ...int) {
	// Look up application name
	app, err := m.env.Dict().GetAppById(m.AppId)
	if app == nil {
		slog.Error(err.Error())
		return
	}

	// Look up command name
	cmd, err := m.env.Dict().GetCmdByCode(m.CmdCode, app)
	if cmd == nil {
		slog.Error(err.Error())
		return
	}

	// Print header fields
	fmt.Printf("Version:  %d\n", m.Version)
	fmt.Printf("Length:   %d\n", m.Length)
	fmt.Printf("AppId:    %d (%s)\n", m.AppId, app.Name)
	fmt.Printf("Flags:    0x%02X", m.Flags)
	if m.Flags != 0 {
		fmt.Print(" (")
		flags := []string{}
		// Iterate through flag bits from R to T
		for flag := m.env.Dict().CmdFlag().R; flag > m.env.Dict().CmdFlag().T; flag >>= 1 {
			if flag&m.Flags != 0 {
				flags = append(flags, m.env.Dict().CmdFlagName(flag))
			}
		}
		fmt.Print(strings.Join(flags, ", "))
		fmt.Print(")")
	}
	fmt.Println()
	fmt.Printf("CmdCode:  %d (%s)\n", m.CmdCode, cmd.Name)
	fmt.Printf("EndToEnd: 0x%08X (%d)\n", m.EndToEnd, m.EndToEnd)
	fmt.Printf("HopByHop: 0x%08X (%d)\n", m.HopByHop, m.HopByHop)
	fmt.Printf("AVPs:\n")
	// Dump all AVPs with 2-space indentation
	for _, avp := range m.avps {
		avp.Dump(2)
	}
	fmt.Println()
}

// Helpers
//
// matchAvp checks if an AVP matches the given identifier.
// The avpId can be:
//   - int/uint32: matches by AVP code
//   - string: matches by AVP name (case-insensitive)
//   - *Avp: matches by pointer identity
func matchAvp(avp *Avp, avpId any) bool {
	switch avpId := avpId.(type) {
	case int:
		return avp.header.Code == uint32(avpId)
	case int32, uint32:
		return avp.header.Code == avpId
	case string:
		return strings.EqualFold(avp.header.Name, avpId)
	case *Avp:
		return avp == avpId
	default:
		return false
	}
}
