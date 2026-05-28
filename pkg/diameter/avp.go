//
// Project: TGDP - Traffic Generator for Diameter Protocol
// Description: Simple tool for testing and debugging the Diameter protocol
//
// Author: Alexander Kefeli <alexander.kefeli@gmail.com>
//
// File: avp.go
// Description: Diameter pkg: AVP handling
//

package diameter

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"reflect"

	"tgdp/pkg/diameter/dict"
	"tgdp/pkg/diameter/diwe"
)

// Types
//

// Avp represents a Diameter Attribute-Value Pair (AVP).
// An AVP is a basic unit of Diameter information, consisting of
// a header (code, flags, vendor ID) and a typed value.
type Avp struct {
	// header contains the AVP definition from the dictionary.
	header dict.Avp
	// value holds the encoded AVP data.
	value *AvpData
	// env is a reference to the Diameter environment for codec access.
	env *Diameter
	// length is the total AVP length in bytes (header + value + padding).
	length uint32
}

// Methods
//

// dict returns a reference to the Diameter dictionary described AVP.
func (avp *Avp) Dict() *dict.Dict {
	return &avp.env.dict
}

// Code returns the AVP code - a unique identifier for this AVP type.
func (avp *Avp) Code() uint32 {
	return avp.header.Code
}

// Name returns the AVP name from the dictionary (e.g., "Session-Id").
func (avp *Avp) Name() string {
	return avp.header.Name
}

// Flags returns the AVP flags byte containing V, M, and P bits.
func (avp *Avp) Flags() uint8 {
	return avp.header.Flags
}

// VendorId returns the Vendor-ID if the V flag is set, otherwise 0.
func (avp *Avp) VendorId() uint32 {
	return avp.header.VndId
}

// Type returns the AVP data type (e.g., Integer32, UTF8String, Grouped).
func (avp *Avp) Type() int {
	return avp.header.Type
}

// Enum returns the enumeration definition for Enumerated-type AVPs,
// or nil if the AVP is not of Enumerated type.
func (avp *Avp) Enum() *dict.Enum {
	return avp.header.Enum
}

// Group returns the group definition for Grouped-type AVPs,
// or nil if the AVP is not of Grouped type.
func (avp *Avp) Group() *dict.Group {
	return avp.header.Group
}

// Data returns the raw AvpData struct containing the encoded value and size.
func (avp *Avp) Data() *AvpData {
	return avp.value
}

// Value returns the decoded Go value of the AVP.
// The type depends on the AVP data type (e.g., string, uint32, []*Avp for grouped).
func (avp *Avp) Value() any {
	if avp.value == nil {
		return nil
	}

	return avp.value.Value
}

// Env returns the Diameter environment which AVP belongs to.
func (avp *Avp) Env() *Diameter {
	return avp.env
}

// Len calculates and returns the total AVP length in bytes.
// This includes the 8-byte header (or 12 bytes if Vendor-Specific),
// plus the value size aligned to a 4-byte boundary.
func (avp *Avp) Len() uint32 {
	avp.length = 8 // base header size
	if avp.IsVendorSpec() {
		avp.length += 4 // additional 4 bytes for Vendor-ID
	}
	if avp.value != nil {
		avp.length += avp.value.Size
	}

	return avp.length
}

// SetValue decodes and sets the AVP value from a Go value.
// It validates that the input type matches the expected Go type for this
// AVP's data type, then uses the appropriate codec to encode the value.
// Returns an error if the value type is invalid or encoding fails.
func (avp *Avp) SetValue(value any) error {
	// Validate input type matches expected Go type for this AVP
	avpGoType, ok := avp.env.dia2go[avp.Type()]
	if !ok {
		return &diwe.ErrUnknownAvpType{Avp: avp.Name(), Type: avp.Type()}
	}
	valueType := reflect.TypeOf(value)
	if valueType != avpGoType.type1 && valueType != avpGoType.type2 {
		return &diwe.ErrInvalidAvpValue{Avp: avp.Name(), Value: value}
	}

	// Use codec to create AvpData from input value
	if codec, exists := avp.env.Codec(avp.Type()); exists {
		v, err := codec.MakeValue(avp, value)
		if err != nil {
			return err
		}
		avp.value = v
	} else {
		return &diwe.ErrUnknownAvpType{Avp: avp.Name(), Type: avp.Type()}
	}

	return nil
}

// Serialize encodes the AVP into the provided buffer in Diameter format.
// The format is: Code (4 bytes) | Flags+Length (4 bytes) | Vendor-ID (optional, 4 bytes) | Data (variable)
// The Flags and Length are packed together in bytes 4-7:
//   - Bits 24-31: Flags byte
//   - Bits 0-23: Length (24-bit unsigned integer)
//
// Returns an error if serialization fails.
func (avp *Avp) Serialize(buf *bytes.Buffer) error {
	bufLen := buf.Len()

	// Write AVP Code (4 bytes, big-endian)
	binaryWrite(buf, binary.BigEndian, avp.Code())

	// Reserve space for Flags+Length (4 bytes)
	lenPos := buf.Len()
	binaryWrite(buf, binary.BigEndian, uint32(0))

	// Write Vendor-ID if V flag is set (4 bytes, big-endian)
	if avp.IsVendorSpec() {
		binaryWrite(buf, binary.BigEndian, avp.VendorId())
	}

	// Serialize the AVP value data using the appropriate codec
	if codec, exists := avp.env.Codec(avp.Type()); exists {
		if err := codec.Serialize(avp, buf); err != nil {
			return err
		}
	} else {
		return &diwe.ErrUnknownAvp{Avp: avp.Name()}
	}

	// Calculate total AVP length (including header)
	avp.length = uint32(buf.Len() - bufLen)

	// Pack flags and length into 4 bytes: [Flags(8bits) | Length(24bits)]
	binary.BigEndian.PutUint32(buf.Bytes()[lenPos:lenPos+4], uint32(avp.Flags())<<24|uint32(avp.length&mask24bits))

	// Write padding bytes to align to 4-byte boundary
	pad := alignTo4(avp.length) - avp.length
	for range pad {
		buf.WriteByte(0)
	}

	return nil
}

// reset clears the AVP state for reuse in a pool.
func (avp *Avp) reset() {
	avp.header = dict.Avp{}
	avp.value = nil
	avp.env = nil
	avp.length = 0
}

// Deserialize decodes an AVP from a byte slice.
// The data should start at the AVP header (Code field).
// It looks up the AVP definition in the dictionary, extracts flags and length,
// reads the Vendor-ID if present, then uses the codec to decode the value data.
//
// Returns the number of bytes consumed (aligned length) and any error that occurred.
func (avp *Avp) Deserialize(data []byte) (uint32, error) {
	if len(data) < 8 {
		return 0, &diwe.ErrAvpTooShort{Len: len(data)}
	}

	// Parse fixed header: Code (4 bytes)
	offset := uint32(4)
	avpCode := binary.BigEndian.Uint32(data[:offset])

	// Parse Flags+Length: Flags in bits 24-31, Length in bits 0-23
	avpFlags := uint8(data[offset])
	avpLength := binary.BigEndian.Uint32(data[offset:offset+4]) & mask24bits
	avpLenAligned := alignTo4(avpLength)
	offset += 4

	// Look up AVP definition in dictionary
	hdr, err := avp.Dict().GetAvp(avpCode)
	if err != nil {
		return avpLenAligned, err
	}
	avp.header = *hdr
	avp.header.Flags = avpFlags
	avp.length = avpLength

	// Parse Vendor-ID if V flag is set
	if avp.IsVendorSpec() {
		avp.header.VndId = binary.BigEndian.Uint32(data[offset : offset+4])
		offset += 4
	}

	// Deserialize value data using the appropriate codec
	if codec, exists := avp.env.codecs[avp.Type()]; exists {
		avp.value = codec.Deserialize(avp, data[offset:], avpLength-offset)
	} else {
		return avpLenAligned, &diwe.ErrUnknownAvpType{Avp: avp.Name(), Type: avp.Type()}
	}

	return avpLenAligned, nil
}

// Copy creates a deep copy of an AVP with copying its value.
func (avp *Avp) Copy() (*Avp, error) {
	if codec, exists := avp.env.Codec(avp.Type()); exists {
		value := codec.CopyValue(avp.Data())

		newAvp := getAvp()
		newAvp.header = dict.Avp{
			Name:  avp.Name(),
			Code:  avp.Code(),
			Flags: avp.Flags(),
			VndId: avp.VendorId(),
			Type:  avp.Type(),
			Enum:  avp.Enum(),
			Group: avp.Group(),
		}
		newAvp.env = avp.env
		newAvp.value = value
		return newAvp, nil
	}

	return nil, &diwe.ErrUnknownAvpType{Avp: avp.Name(), Type: avp.Type()}
}

// IsVendorSpec returns true if the Vendor-Specific bit (V) is set in flags.
// When true, the AVP includes a Vendor-ID field and is defined by a specific vendor.
func (avp *Avp) IsVendorSpec() bool {
	return (avp.Flags()&avp.Dict().AvpFlag().V != 0)
}

// IsMandatory returns true if the Mandatory bit (M) is set in flags.
// Mandatory AVPs must be present and understood by the receiving peer.
func (avp *Avp) IsMandatory() bool {
	return (avp.Flags()&avp.Dict().AvpFlag().M != 0)
}

// IsProtected returns true if the Protected bit (P) is set in flags.
// Protected AVPs must be encrypted for end-to-end security.
func (avp *Avp) IsProtected() bool {
	return (avp.header.Flags&avp.Dict().AvpFlag().P != 0)
}

// IsGrouped returns true if the AVP data type is Grouped.
// Grouped AVPs contain nested AVPs as their value.
func (avp *Avp) IsGrouped() bool {
	return avp.Type() == avp.Dict().AvpDataType().Grouped
}

// Print outputs the AVP to stdout in human-readable format.
// This is an alias for Dump with no indentation shift.
func (avp *Avp) Print(shift ...int) {
	avp.Dump(shift...)
}

// Dump outputs the AVP to stdout in human-readable format.
// For grouped AVPs, recursively dumps all nested member AVPs.
// Output format: "Name (Code): Value" or "Name (Code):" followed by nested AVPs.
func (avp *Avp) Dump(shift ...int) {
	indent := 0
	if len(shift) > 0 {
		indent = shift[0]
	}

	// Print indentation
	for range indent {
		fmt.Print(" ")
	}

	// Print AVP name and code
	if len(avp.Name()) > 0 {
		fmt.Printf("%s (%d): ", avp.Name(), avp.Code())
	} else {
		fmt.Printf("Unknown <%d>", avp.Code())
	}

	// Handle grouped AVPs - recursively dump members via AvpStore
	if avp.IsGrouped() {
		fmt.Println()
		for _, member := range avp.Value().([]*Avp) {
			member.Dump(indent + 2)
		}
		return
	}

	if codec, exists := avp.env.Codec(avp.Type()); exists {
		fmt.Printf("%s", codec.ToText(avp))
	}

	fmt.Println()
}

// Helpers
//

func binaryWrite(b *bytes.Buffer, o binary.ByteOrder, a any) {
	switch v := a.(type) {
	case uint32:
		var buf [4]byte
		binary.BigEndian.PutUint32(buf[:], v)
		b.Write(buf[:])
		return
	}
	if err := binary.Write(b, o, a); err != nil {
		panic(err)
	}
}

// alignTo4 rounds up a length to the nearest 4-byte boundary.
// Diameter AVPs must be padded to a multiple of 4 bytes.
func alignTo4(n uint32) uint32 {
	return ((n + 3) &^ 3)
}
