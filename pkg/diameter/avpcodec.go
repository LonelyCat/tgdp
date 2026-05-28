//
// Project: TGDP - Traffic Generator for Diameter Protocol
// Description: Simple tool for testing and debugging the Diameter protocol
//
// Author: Alexander Kefeli <alexander.kefeli@gmail.com>
//
// File: avpcodec.go
// Description: Diameter pkg: AVP data processing functions
//

package diameter

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"net"
	"net/netip"
	"slices"
	"strconv"
	"strings"
	"time"

	"tgdp/pkg/diameter/diwe"
)

// Consts
//

const (
	maskHigh4bits = 0xF0
	maskLow4bits  = 0x0F

	zeroChar = 0x30 // '0'
	fChar    = 0x3F // 'F'

	addrIPv4 = 1
	addrIPv6 = 2
)

// Types
//

// AvpCodecs is a map of AVP type IDs to their codec functions.
// This allows the system to encode/decode different AVP data types dynamically.
type AvpCodecs map[int]CodecFuncs

// CodecFuncs holds the three functions needed to work with an AVP type:
// makeValue creates an AvpData from a Go value, serialize writes to wire format,
// and deserialize parses from wire format.
type CodecFuncs struct {
	MakeValue   MakeValueFn
	Serialize   SerializeFn
	Deserialize DeserializeFn
	CopyValue   CopyValueFn
	ToText      ToTextFn
}

// Codec Function Types
//

// MakeValueFn is a function that creates an AvpData from a Go value.
// It validates the input type and performs any necessary encoding.
type MakeValueFn func(*Avp, any) (*AvpData, error)

// SerializeFn is a function that serializes an AVP value to wire format.
// It writes the encoded data to the provided buffer.
type SerializeFn func(*Avp, *bytes.Buffer) error

// DeserializeFn is a function that deserializes an AVP value from wire format.
// It parses the raw bytes and returns an AvpData with the decoded value.
type DeserializeFn func(*Avp, []byte, uint32) *AvpData

// CopyValueFn is a function that copy value from an AVP.
// It accepts source AVP data as parameter and returns an AvpData with the copied value.
type CopyValueFn func(*AvpData) *AvpData

// ToTextFn is a function represent AVP as text.
// It accepts source AVP data as parameter and returns an AvpData with the copied value.
type ToTextFn func(*Avp) string

// specEncodingFn is a function that performs special encoding for specific AVP codes.
// These are custom encoders for vendor-specific or specialized AVP formats.
type specEncodingFn func(data []byte) (*AvpData, error)

// specToTextFn is a function that performs special conversation to text for specific AVP codes.
// These are custom decoders for vendor-specific or specialized AVP formats.
type specToTextFn func(data []byte) string

// Variables
//

// specEncoders maps AVP codes to their special encoding functions.
// Currently only contains the PLMN (Visited-PLMN-Id) encoder.
var (
	specEncoders = map[uint32]specEncodingFn{
		1407: encodePLMN, // Visited-PLMN-Id
	}

	specToText = map[uint32]specToTextFn{
		701:  toTextMSISDN, // MSISDN
		1407: toTextPLMN,   // Visited-PLMN-Id
	}
)

// Make Value Functions
//
// These functions convert Go values into AvpData for serialization.

// mkvOctetString creates AvpData from a string or int value.
// It handles special encoders for specific AVP codes and can convert
// numeric strings to packed BCD format for compatibility.
func mkvOctetString(avp *Avp, value any) (*AvpData, error) {
	rawBytes, err := func() ([]byte, error) {
		switch v := (value).(type) {
		case int:
			return []byte(strconv.Itoa(v)), nil
		case string:
			return []byte(v), nil
		default:
			return nil, &diwe.ErrInvalidAvpValue{Avp: avp, Value: v}
		}
	}()

	if err != nil {
		return nil, err
	}

	// Check if this AVP has a special encoder (e.g., PLMN)
	if osse, exists := specEncoders[avp.Code()]; exists {
		return osse(rawBytes)
	}

	// Check if the string contains only numeric characters
	isNumeric := true
	for _, r := range rawBytes {
		if r < '0' || r > '9' {
			isNumeric = false
			break
		}
	}

	// If numeric, pack as BCD (two digits per byte)
	encoded := rawBytes
	if isNumeric {
		bytes := make([]byte, (len(rawBytes)+1)/2)
		i := 0
		for ; i < len(rawBytes); i++ {
			if i&1 == 0 {
				bytes[i/2] += (rawBytes[i] - '0')
			} else {
				bytes[i/2] += (rawBytes[i] - '0') << 4
			}
		}
		// Add filler for odd-length strings
		if i&1 != 0 {
			bytes[i/2] += maskHigh4bits
		}
		encoded = bytes
	}

	return &AvpData{
			Value: encoded,
			Size:  uint32(len(encoded)),
		},
		nil
}

// mkvInteger32 creates AvpData from an int32 value.
func mkvInteger32(avp *Avp, value any) (*AvpData, error) {
	v, ok := (value).(int32)
	if !ok {
		return nil, &diwe.ErrInvalidAvpValue{Avp: avp.Name(), Value: value}
	}

	return &AvpData{
			Value: v,
			Size:  4,
		},
		nil
}

// mkvUnsigned32 creates AvpData from a uint32 value.
func mkvUnsigned32(avp *Avp, value any) (*AvpData, error) {
	v, ok := (value).(uint32)
	if !ok {
		return nil, &diwe.ErrInvalidAvpValue{Avp: avp.Name(), Value: value}
	}

	return &AvpData{
			Value: v,
			Size:  4,
		},
		nil
}

// mkvInteger64 creates AvpData from an int64 value.
func mkvInterger64(avp *Avp, value any) (*AvpData, error) {
	v, ok := (value).(int64)
	if !ok {
		return nil, &diwe.ErrInvalidAvpValue{Avp: avp.Name(), Value: value}
	}

	return &AvpData{
			Value: v,
			Size:  8,
		},
		nil
}

// mkvUnsigned64 creates AvpData from a uint64 value.
func mkvUnsigned64(avp *Avp, value any) (*AvpData, error) {
	v, ok := (value).(uint64)
	if !ok {
		return nil, &diwe.ErrInvalidAvpValue{Avp: avp.Name(), Value: value}
	}

	return &AvpData{
			Value: v,
			Size:  8,
		},
		nil
}

// mkvFloat32 creates AvpData from a float32 value.
func mkvFloat32(avp *Avp, value any) (*AvpData, error) {
	v, ok := (value).(float32)
	if !ok {
		return nil, &diwe.ErrInvalidAvpValue{Avp: avp.Name(), Value: value}
	}

	return &AvpData{
			Value: v,
			Size:  4,
		},
		nil
}

// mkvFloat64 creates AvpData from a float64 value.
func mkvFloat64(avp *Avp, value any) (*AvpData, error) {
	v, ok := (value).(float64)
	if !ok {
		return nil, &diwe.ErrInvalidAvpValue{Avp: avp.Name(), Value: value}
	}

	return &AvpData{
			Value: v,
			Size:  8,
		},
		nil
}

// mkvIpAddress creates AvpData from an IP address string.
// Supports both IPv4 (returns size 6) and IPv6 (returns size 18).
// The size includes a 2-byte address type prefix (1 for IPv4, 2 for IPv6).
func mkvIpAddress(avp *Avp, value any) (*AvpData, error) {
	ip, size, err := func() (netip.Addr, uint32, error) {
		// Handle netip.Addr directly
		if v, ok := (value).(netip.Addr); ok {
			// IPv4: 2-byte prefix + 4 bytes = 6 total
			// IPv6: 2-byte prefix + 16 bytes = 18 total
			if v.Is4() {
				return v, 6, nil
			} else {
				return v, 18, nil
			}
		}

		// Handle net.IP (legacy Go type)
		if v, ok := (value).(net.IP); ok {
			ip, err := netip.ParseAddr(v.String())
			if err != nil {
				return netip.Addr{}, 0, &diwe.ErrInvalidAvpValue{Avp: avp.Name(), Value: v}
			}
			if ip.Is4() {
				return ip, 6, nil
			} else {
				return ip, 18, nil
			}
		}

		// Handle string (current behavior)
		if v, ok := (value).(string); ok {
			ip, err := netip.ParseAddr(v)
			if err != nil {
				return netip.Addr{}, 0, &diwe.ErrInvalidAvpValue{Avp: avp.Name(), Value: v}
			}
			if ip.Is4() {
				return ip, 6, nil
			} else {
				return ip, 18, nil
			}
		}

		return netip.Addr{}, 0, &diwe.ErrInvalidAvpValue{Avp: avp.Name(), Value: value}
	}()

	if err != nil {
		return nil, err
	}

	return &AvpData{
		Value: ip,
		Size:  size,
	}, nil
}

// mkvTime creates AvpData from a time.Time or RFC3339 string.
// Diameter Time type uses seconds since Unix epoch (32-bit).
func mkvTime(avp *Avp, value any) (*AvpData, error) {
	t, err := func() (int64, error) {
		if v, ok := (value).(time.Time); ok {
			return v.Unix(), nil
		}

		if v, ok := (value).(string); ok {
			if t, err := time.Parse(time.RFC3339, v); err == nil {
				return t.Unix(), nil
			}
		}

		return 0, &diwe.ErrInvalidAvpValue{Avp: avp.Name(), Value: value}
	}()

	if err != nil {
		return nil, err
	}

	return &AvpData{
			Value: t,
			Size:  4,
		},
		nil
}

// mkvUTF8String creates AvpData from a UTF-8 string value.
func mkvUTF8String(avp *Avp, value any) (*AvpData, error) {
	v, ok := (value).(string)
	if !ok {
		return nil, &diwe.ErrInvalidAvpValue{Avp: avp.Name(), Value: value}
	}

	return &AvpData{
			Value: v,
			Size:  uint32(len(v)),
		},
		nil
}

// mkvEnumerated creates AvpData from an enumerated value.
// Accepts either an int (used as the enum code) or a string (looked up by name).
// Performs case-insensitive name matching.
func mkvEnumerated(avp *Avp, value any) (*AvpData, error) {
	v, err := func() (int32, error) {
		switch v := (value).(type) {
		case int:
			return int32(v), nil
		case string:
			// Look up enum item by name using dictionary cache
			return avp.Dict().GetEnumCode(avp.Code(), v)
		}
		return 0, &diwe.ErrUnknownEnumItem{Avp: avp.Name(), Value: value}
	}()

	if err != nil {
		return nil, err
	}

	return &AvpData{
			Value: v,
			Size:  4,
		},
		nil
}

// mkvGrouped creates AvpData from an AvpStore containing nested AVPs.
func mkvGrouped(avp *Avp, value any) (*AvpData, error) {
	v, ok := (value).([]*Avp)
	if !ok {
		return nil, &diwe.ErrInvalidAvpValue{Avp: avp.Name(), Value: value}
	}

	members := avp.Value()
	if members == nil {
		members = make([]*Avp, 0, len(v))
	}
	members = append(members.([]*Avp), v...)

	return &AvpData{
			Value: members,
			Size:  0,
		},
		nil
}

// Serialize Functions
//
// These functions write AVP values to wire format in big-endian byte order.

// serOctetString writes raw bytes to the buffer.
func serOctetString(avp *Avp, buf *bytes.Buffer) error {
	if v, ok := avp.Value().([]byte); ok {
		return binary.Write(buf, binary.BigEndian, v)
	}
	return &diwe.ErrInvalidAvpValue{Avp: avp, Value: avp.Value()}
}

// serInteger32 writes a 32-bit signed integer in big-endian order.
func serInteger32(avp *Avp, buf *bytes.Buffer) error {
	if v, ok := avp.Value().(int32); ok {
		return binary.Write(buf, binary.BigEndian, v)
	}
	return &diwe.ErrInvalidAvpValue{Avp: avp, Value: avp.Value()}
}

// serUnsigned32 writes a 32-bit unsigned integer in big-endian order.
func serUnsigned32(avp *Avp, buf *bytes.Buffer) error {
	if v, ok := avp.Value().(uint32); ok {
		return binary.Write(buf, binary.BigEndian, v)
	}
	return &diwe.ErrInvalidAvpValue{Avp: avp, Value: avp.Value()}
}

// serInteger64 writes a 64-bit signed integer in big-endian order.
func serInteger64(avp *Avp, buf *bytes.Buffer) error {
	if v, ok := avp.Value().(int64); ok {
		return binary.Write(buf, binary.BigEndian, v)
	}
	return &diwe.ErrInvalidAvpValue{Avp: avp, Value: avp.Value()}
}

// serUnsigned64 writes a 64-bit unsigned integer in big-endian order.
func serUnsigned64(avp *Avp, buf *bytes.Buffer) error {
	if v, ok := avp.Value().(uint64); ok {
		return binary.Write(buf, binary.BigEndian, v)
	}
	return &diwe.ErrInvalidAvpValue{Avp: avp, Value: avp.Value()}
}

// serFloat32 writes a 32-bit floating point in big-endian IEEE 754 format.
func serFloat32(avp *Avp, buf *bytes.Buffer) error {
	if v, ok := avp.Value().(float32); ok {
		return binary.Write(buf, binary.BigEndian, v)
	}
	return &diwe.ErrInvalidAvpValue{Avp: avp, Value: avp.Value()}
}

// serFloat64 writes a 64-bit floating point in big-endian IEEE 754 format.
func serFloat64(avp *Avp, buf *bytes.Buffer) error {
	if v, ok := avp.Value().(float64); ok {
		return binary.Write(buf, binary.BigEndian, v)
	}
	return &diwe.ErrInvalidAvpValue{Avp: avp, Value: avp.Value()}
}

// serIpAddress writes an IP address with a 2-byte type prefix.
// Type 1 = IPv4 (4 bytes follow), Type 2 = IPv6 (16 bytes follow).
func serIpAddress(avp *Avp, buf *bytes.Buffer) error {
	if addr, ok := avp.Value().(netip.Addr); ok {
		var addrType uint16
		if addr.Is4() {
			addrType = 1
		} else {
			addrType = 2
		}

		if err := binary.Write(buf, binary.BigEndian, addrType); err != nil {
			return err
		} else {
			return binary.Write(buf, binary.BigEndian, addr.AsSlice())
		}
	}
	return &diwe.ErrInvalidAvpValue{Avp: avp, Value: avp.Value()}
}

// serTime writes a 32-bit unsigned timestamp (seconds since Unix epoch).
func serTime(avp *Avp, buf *bytes.Buffer) error {
	if v, ok := avp.Value().(uint32); ok {
		return binary.Write(buf, binary.BigEndian, v)
	}
	return &diwe.ErrInvalidAvpValue{Avp: avp, Value: avp.Value()}
}

// serUTF8String writes a UTF-8 string as raw bytes.
func serUTF8String(avp *Avp, buf *bytes.Buffer) error {
	if v, ok := avp.Value().(string); ok {
		_, err := buf.WriteString(v)
		return err
	}
	return &diwe.ErrInvalidAvpValue{Avp: avp, Value: avp.Value()}
}

// serEnumerated writes a 32-bit signed integer for enumerated values.
func serEnumerated(avp *Avp, buf *bytes.Buffer) error {
	if v, ok := avp.Value().(int32); ok {
		return binary.Write(buf, binary.BigEndian, v)
	}
	return &diwe.ErrInvalidAvpValue{Avp: avp, Value: avp.Value()}
}

// serGrouped serializes all nested AVPs from an AvpStore recursively.
func serGrouped(avp *Avp, buf *bytes.Buffer) error {
	for _, member := range avp.Value().([]*Avp) {
		if err := member.Serialize(buf); err != nil {
			return err
		}
	}

	return nil
}

// Deserialize Functions
//
// These functions parse AVP values from wire format in big-endian byte order.

// desOctetString returns raw bytes as the value.
func desOctetString(avp *Avp, data []byte, size uint32) *AvpData {
	return &AvpData{
		Value: data[:size],
		Size:  size,
	}
}

// desInteger32 reads a 32-bit signed integer.
func desInteger32(avp *Avp, data []byte, size uint32) *AvpData {
	return &AvpData{
		Value: int32(binary.BigEndian.Uint32(data[:4])),
		Size:  4,
	}
}

// desUnsigned32 reads a 32-bit unsigned integer.
func desUnsigned32(avp *Avp, data []byte, size uint32) *AvpData {
	return &AvpData{
		Value: binary.BigEndian.Uint32(data[:4]),
		Size:  4,
	}
}

// desInteger64 reads a 64-bit signed integer.
func desInteger64(avp *Avp, data []byte, size uint32) *AvpData {
	return &AvpData{
		Value: int64(binary.BigEndian.Uint64(data[:8])),
		Size:  8,
	}
}

// desUnsigned64 reads a 64-bit unsigned integer.
func desUnsigned64(avp *Avp, data []byte, size uint32) *AvpData {
	return &AvpData{
		Value: binary.BigEndian.Uint64(data[:8]),
		Size:  8,
	}
}

// desFloat32 reads a 32-bit floating point in IEEE 754 format.
func desFloat32(avp *Avp, data []byte, size uint32) *AvpData {
	return &AvpData{
		Value: math.Float32frombits(binary.BigEndian.Uint32(data[:4])),
		Size:  4,
	}
}

// desFloat64 reads a 64-bit floating point in IEEE 754 format.
func desFloat64(avp *Avp, data []byte, size uint32) *AvpData {
	return &AvpData{
		Value: math.Float64frombits(binary.BigEndian.Uint64(data[:8])),
		Size:  8,
	}
}

// desIpAddress reads an IP address from the 2-byte type prefix + address data.
func desIpAddress(avp *Avp, data []byte, size uint32) *AvpData {
	var (
		addr netip.Addr
		ok   bool
	)
	// Read address type: 1 = IPv4, 2 = IPv6
	addrType := binary.BigEndian.Uint16(data[:2])

	switch addrType {
	case addrIPv4:
		addr, ok = netip.AddrFromSlice(data[2:6])
		size = 4
	case addrIPv6:
		addr, ok = netip.AddrFromSlice(data[2:18])
		size = 16
	}

	if !ok {
		return nil
	}

	return &AvpData{
		Value: addr,
		Size:  size,
	}
}

// desTime reads a 32-bit unsigned timestamp (seconds since Unix epoch).
func desTime(avp *Avp, data []byte, size uint32) *AvpData {
	return &AvpData{
		Value: binary.BigEndian.Uint32(data[:4]),
		Size:  4,
	}
}

// desUTF8String reads a UTF-8 string from raw bytes.
func desUTF8String(avp *Avp, data []byte, size uint32) *AvpData {
	return &AvpData{
		Value: string(data[:size]),
		Size:  size,
	}
}

// desEnumerated reads a 32-bit signed integer for enumerated values.
func desEnumerated(avp *Avp, data []byte, size uint32) *AvpData {
	return &AvpData{
		Value: int32(binary.BigEndian.Uint32(data[:4])),
		Size:  4,
	}
}

// desGrouped recursively deserializes nested AVPs from the data.
// It parses each nested AVP sequentially until the total size is consumed.
func desGrouped(avp *Avp, data []byte, size uint32) *AvpData {
	members := make([]*Avp, 0)
	offset := uint32(0)
	for offset < size {
		member := &Avp{env: avp.env}
		o, err := member.Deserialize(data[offset:])
		if err != nil {
			break
		}

		offset += o
		members = append(members, member)
	}

	return &AvpData{
		Value: members,
		Size:  offset,
	}
}

// Copy Value Functions
//
// These functions create a copy of AVP values.

// cpvOctetString creates a copy of raw bytes as the value.
func cpvOctetString(srcData *AvpData) *AvpData {
	return &AvpData{
		Value: srcData.Value.([]byte),
		Size:  srcData.Size,
	}
}

// cpvInteger32 creates a copy of a 32-bit signed integer as the value.
func cpvInteger32(srcData *AvpData) *AvpData {
	return &AvpData{
		Value: srcData.Value.(int32),
		Size:  srcData.Size,
	}
}

// cpvUnsigned32 creates a copy of a 32-bit unsigned integer as the value.
func cpvUnsigned32(srcData *AvpData) *AvpData {
	return &AvpData{
		Value: srcData.Value.(uint32),
		Size:  srcData.Size,
	}
}

// cpvInteger64 creates a copy of a 64-bit signed integer as the value.
func cpvInteger64(srcData *AvpData) *AvpData {
	return &AvpData{
		Value: srcData.Value.(int64),
		Size:  srcData.Size,
	}
}

// cpvUnsigned64 creates a copy of a 64-bit unsigned integer as the value.
func cpvUnsigned64(srcData *AvpData) *AvpData {
	return &AvpData{
		Value: srcData.Value.(uint64),
		Size:  srcData.Size,
	}
}

// cpvFloat32 creates a copy of a 32-bit floating point as the value.
func cpvFloat32(srcData *AvpData) *AvpData {
	return &AvpData{
		Value: srcData.Value.(float32),
		Size:  srcData.Size,
	}
}

// cpvFloat64 creates a copy of a 64-bit floating point as the value.
func cpvFloat64(srcData *AvpData) *AvpData {
	return &AvpData{
		Value: srcData.Value.(float64),
		Size:  srcData.Size,
	}
}

// cpvIpAddress creates a copy of an IP address as the value.
func cpvIpAddress(srcData *AvpData) *AvpData {
	return &AvpData{
		Value: srcData.Value.(netip.Addr),
		Size:  srcData.Size,
	}
}

// cpvTime creates a copy of a timestamp as the value.
func cpvTime(srcData *AvpData) *AvpData {
	return &AvpData{
		Value: srcData.Value.(uint32),
		Size:  srcData.Size,
	}
}

// cpvUTF8String creates a copy of a UTF-8 string as the value.
func cpvUTF8String(srcData *AvpData) *AvpData {
	return &AvpData{
		Value: srcData.Value.(string),
		Size:  srcData.Size,
	}
}

// cpvEnumerated creates a copy of an enumerated value as the value.
func cpvEnumerated(srcData *AvpData) *AvpData {
	return &AvpData{
		Value: srcData.Value.(int32),
		Size:  srcData.Size,
	}
}

// cpvGrouped creates a copy of a grouped AVP (AvpStore) as the value.
func cpvGrouped(srcData *AvpData) *AvpData {
	return &AvpData{
		Value: slices.Clone(srcData.Value.([]*Avp)),
		Size:  srcData.Size,
	}
}

// To Text Functions
//
// These functions convert AVP values to human-readable text representation.

// txtOctetString converts raw bytes to a hex string representation.
func txtOctetString(avp *Avp) string {
	if v, ok := avp.Value().([]byte); ok {
		// Check if this AVP has a special decoder (e.g., PLMN)
		if ostt, exists := specToText[avp.Code()]; exists {
			return ostt(v)
		}

		return fmt.Sprintf("%x", v)
	}
	return ""
}

// txtInteger32 converts a 32-bit signed integer to string.
func txtInteger32(avp *Avp) string {
	if v, ok := avp.Value().(int32); ok {
		return strconv.FormatInt(int64(v), 10)
	}
	return ""
}

// txtUnsigned32 converts a 32-bit unsigned integer to string.
func txtUnsigned32(avp *Avp) string {
	if v, ok := avp.Value().(uint32); ok {
		return strconv.FormatUint(uint64(v), 10)
	}
	return ""
}

// txtInteger64 converts a 64-bit signed integer to string.
func txtInteger64(avp *Avp) string {
	if v, ok := avp.Value().(int64); ok {
		return strconv.FormatInt(v, 10)
	}
	return ""
}

// txtUnsigned64 converts a 64-bit unsigned integer to string.
func txtUnsigned64(avp *Avp) string {
	if v, ok := avp.Value().(uint64); ok {
		return strconv.FormatUint(v, 10)
	}
	return ""
}

// txtFloat32 converts a 32-bit floating point to string.
func txtFloat32(avp *Avp) string {
	if v, ok := avp.Value().(float32); ok {
		return strconv.FormatFloat(float64(v), 'f', -1, 32)
	}
	return ""
}

// txtFloat64 converts a 64-bit floating point to string.
func txtFloat64(avp *Avp) string {
	if v, ok := avp.Value().(float64); ok {
		return strconv.FormatFloat(v, 'f', -1, 64)
	}
	return ""
}

// txtIpAddress converts an IP address to string representation.
func txtIpAddress(avp *Avp) string {
	if v, ok := avp.Value().(netip.Addr); ok {
		return v.String()
	}
	return ""
}

// txtTime converts a timestamp (seconds since Unix epoch) to RFC3339 string.
func txtTime(avp *Avp) string {
	if v, ok := avp.Value().(uint32); ok {
		t := time.Unix(int64(v), 0)
		return t.Format(time.RFC3339)
	}
	return ""
}

// txtUTF8String returns the UTF-8 string value as-is.
func txtUTF8String(avp *Avp) string {
	if v, ok := avp.Value().(string); ok {
		return v
	}
	return ""
}

// txtEnumerated converts an enumerated value to its name if available,
// otherwise returns the numeric value as string.
func txtEnumerated(avp *Avp) string {
	if v, ok := avp.Value().(int32); ok {
		// Try to find enum name
		for _, item := range avp.Enum().Items {
			if item.Code == v {
				return fmt.Sprintf("%s (%d)", item.Name, v)
			}
		}
		// Return numeric value if name not found
		return strconv.FormatInt(int64(v), 10)
	}
	return ""
}

// txtGrouped converts a grouped AVP to a string representation of nested AVPs.
func txtGrouped(avp *Avp) string {
	if members, ok := avp.Value().([]*Avp); ok {
		var b strings.Builder
		b.WriteString("{")
		first := true
		for _, member := range members {
			if !first {
				b.WriteString(", ")
			}
			first = false
			b.WriteString(member.Name())
			b.WriteString(": ")
			txtFn := avp.env.codecs[member.Type()].ToText
			b.WriteString(txtFn(member))
		}
		b.WriteString("}")
		return b.String()
	}
	return ""
}

// Special Encoding Functions
//

// encodePLMN encodes a PLMN (Public Land Mobile Network) identifier.
// The input is a string of 5 or 6 digits (MCC + MNC).
// Output is 3 bytes of packed BCD format:
//   - Byte 0: MCC[0] | MCC[1]<<4
//   - Byte 1: MCC[2] | MNC[0]<<4 (or 0xF if 5 digits)
//   - Byte 2: MNC[1] | MNC[2]<<4 (or only MNC[1] if 5 digits)
func encodePLMN(plmn []byte) (*AvpData, error) {
	if len(plmn) != 5 && len(plmn) != 6 {
		return nil, &diwe.ErrInvalidValue{Value: plmn}
	}

	bytes := make([]byte, 3)
	// Pack first two digits (MCC)
	bytes[0] = (plmn[0] - zeroChar) | (plmn[1]-zeroChar)<<4
	if len(plmn) == 6 {
		// 6 digits: MCC[2] + all three MNC digits
		bytes[1] = (plmn[2] - zeroChar) | (plmn[3]-zeroChar)<<4
		bytes[2] = (plmn[4] - zeroChar) | (plmn[5]-zeroChar)<<4
	} else {
		// 5 digits: MCC[2] + filler + MNC[1-2]
		bytes[1] = (plmn[2] - zeroChar) | (maskLow4bits << 4)
		bytes[2] = (plmn[3] - zeroChar) | (plmn[4]-zeroChar)<<4
	}
	return &AvpData{
			Value: bytes,
			Size:  uint32(len(bytes)),
		},
		nil
}

// Special Decoding Functions
//

// toTextPLMN decodes a PLMN identifier to text string.
// The input is a array of 3 bytes.
func toTextPLMN(data []byte) string {
	if len(data) != 3 {
		return ""
	}

	b0, b1, b2 := data[0], data[1], data[2]

	// Decode byte 0: Contains MCC digits 1 and 2
	// e.g., "25" in 0x52
	d1 := (b0 & maskLow4bits) + zeroChar
	d2 := ((b0 >> 4) & maskLow4bits) + zeroChar

	// Decode byte 1: Contains MCC digit 3 (or MNC digit 1 in 6-digit case)
	// If high nibble is 0xF, it indicates 5-digit mode (filler), so we take the low nibble.
	// Otherwise, we take the high nibble.
	var d3 byte
	if b1&maskHigh4bits == maskHigh4bits {
		// 5-digit PLMN
		d3 = (b1 & maskLow4bits) + zeroChar
	} else {
		// 6-digit PLMN
		d3 = ((b1 >> 4) & maskLow4bits) + zeroChar
	}

	// Decode byte 2: Contains MNC digits 4 and 5
	d4 := (b2 & maskLow4bits) + zeroChar
	d5 := ((b2 >> 4) & maskLow4bits) + zeroChar

	return string([]byte{d1, d2, d3, d4, d5})
}

// toTextMSISDN decodes a MSISDN number to text string.
// The input is a array of bytes.
func toTextMSISDN(data []byte) string {
	i := 0
	j := 0
	b := make([]byte, len(data)*2)
	for ; i < len(data)-1; i++ {
		b[j] = (data[i] & maskLow4bits) + zeroChar
		b[j+1] = ((data[i] >> 4) & maskLow4bits) + zeroChar
		j += 2
	}
	b[j] = (data[i] & maskLow4bits) + zeroChar
	j++
	if f := ((data[i] >> 4) & maskLow4bits) + zeroChar; f != fChar {
		b[j] = f
		j++
	}

	return string(b[:j])
}
