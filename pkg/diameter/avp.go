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
	"log/slog"
	"math"
	"net"
	"net/netip"
	"reflect"
	"slices"
	"strings"
)

// -- Types
// --
type avpGoType struct {
	type1 reflect.Type
	type2 reflect.Type
}

// -- Variables
// --
var (
	dia2Go map[int]avpGoType
)

// -- Functions
// --
func NewAvp(name string, code uint32, flags uint8, vendor_id uint32, datatype int) *Avp {
	avp := new(Avp)
	avp.Name = name
	avp.Code = code
	avp.Flags = flags
	avp.VndId = vendor_id
	avp.Type = datatype
	if vendor_id != 0 {
		avp.Flags |= Dict.AvpFlagV()
	}
	return avp
}

// -- Methods
// --
func (avp *Avp) Len() uint32 {
	avp.Length = 8
	if avp.Data != nil {
		avp.Length += avp.Data.Size
	}
	if avp.IsVendorSpec() {
		avp.Length += 4
	}
	return avp.Length
}

func (avp *Avp) GetValue() any {
	return avp.Data.Value
}

func (avp *Avp) SetValue(value *any) error {
	if avpData, err := avp.MakeValue(value); err != nil {
		return err
	} else {
		avp.Data = avpData
		avp.Len()
		return nil
	}
}

func (avp *Avp) MakeValue(value *any) (*AvpData, error) {
	if value != nil {
		avpGoType := dia2Go[avp.Type]
		valueType := reflect.TypeOf(*value)
		if valueType != avpGoType.type1 && valueType != avpGoType.type2 {
			return nil, &ErrInvalidAvpValue{avp, *value}
		}
	}

	avpData := new(AvpData)
	switch avp.Type {
	case Dict.AvpTypeOctetString():
		if bytes, err := avp.encodeOctetString(*value); err != nil {
			return nil, err
		} else {
			avpData.Value = bytes
			avpData.Size = uint32(len(bytes))
		}
	case Dict.AvpTypeInteger32():
		avpData.Value = (*value).(int32)
		avpData.Size = 4
	case Dict.AvpTypeInteger64():
		avpData.Value = (*value).(int64)
		avpData.Size = 8
	case Dict.AvpTypeUnsigned32():
		avpData.Value = (*value).(uint32)
		avpData.Size = 4
	case Dict.AvpTypeUnsigned64():
		avpData.Value = (*value).(uint64)
		avpData.Size = 8
	case Dict.AvpTypeFloat32():
		avpData.Value = (*value).(float64)
		avpData.Size = 4
	case Dict.AvpTypeFloat64():
		avpData.Value = (*value).(float64)
		avpData.Size = 8
	case Dict.AvpTypeAddress():
		if ip, size, err := avp.encodeAddr(*value); err != nil {
			return nil, err
		} else {
			avpData.Value = ip
			avpData.Size = size
		}
	case Dict.AvpTypeTime():
		if t, err := avp.encodeTime(*value); err != nil {
			return nil, err
		} else {
			avpData.Value = t
			avpData.Size = 4
		}
	case Dict.AvpTypeUTF8String():
		avpData.Size = uint32(len((*value).(string)))
		avpData.Value = (*value).(string)
	case Dict.AvpTypeIdentity():
		avpData.Size = uint32(len((*value).(string)))
		avpData.Value = (*value).(string)
	case Dict.AvpTypeURI():
		avpData.Size = uint32(len((*value).(string)))
		avpData.Value = (*value).(string)
	case Dict.AvpTypeIPFilterRule():
		avpData.Size = uint32(len((*value).(string)))
		avpData.Value = (*value).(string)
	case Dict.AvpTypeQoSFilterRule():
		avpData.Size = uint32(len((*value).(string)))
		avpData.Value = (*value).(string)
	case Dict.AvpTypeEnumerated():
		if code, err := avp.enumItemToCode(*value); err != nil {
			return nil, err
		} else {
			avpData.Value = code
			avpData.Size = 4
		}
	case Dict.AvpTypeGrouped():
		avpData.Value = (*value).(*AvpDataStore)
		avpData.Size = 0
	default:
		return nil, &ErrUnknownAvpType{Avp: avp}
	}

	return avpData, nil
}

func (avp *Avp) AddMember(name string, required bool, max int) {
	member := &AvpRule{Name: name, Required: required, Max: max}
	avp.Group.Members = append(avp.Group.Members, member)
}

func (avp *Avp) RemoveMember(name string) {
	for i, member := range avp.Group.Members {
		if strings.EqualFold(member.Name, name) {
			avp.Group.Members = slices.Delete(avp.Group.Members, i, i+1)
		}
	}
}

func (avp *Avp) Serialize(buf *bytes.Buffer) error {
	binary.Write(buf, binary.BigEndian, avp.Code)
	binary.Write(buf, binary.BigEndian, uint32(avp.Flags)<<24|uint32(avp.Length&0x00FFFFFF))
	if avp.IsVendorSpec() && avp.VndId != 0 {
		binary.Write(buf, binary.BigEndian, avp.VndId)
	}

	switch avp.Type {
	case Dict.AvpTypeOctetString():
		binary.Write(buf, binary.BigEndian, (avp.Data.Value).([]byte))
	case Dict.AvpTypeInteger32():
		binary.Write(buf, binary.BigEndian, (avp.Data.Value).(int32))
	case Dict.AvpTypeInteger64():
		binary.Write(buf, binary.BigEndian, (avp.Data.Value).(int64))
	case Dict.AvpTypeUnsigned32():
		binary.Write(buf, binary.BigEndian, (avp.Data.Value).(uint32))
	case Dict.AvpTypeUnsigned64():
		binary.Write(buf, binary.BigEndian, (avp.Data.Value).(uint64))
	case Dict.AvpTypeFloat32():
		binary.Write(buf, binary.BigEndian, (avp.Data.Value).(float32))
	case Dict.AvpTypeFloat64():
		binary.Write(buf, binary.BigEndian, (avp.Data.Value).(float64))
	case Dict.AvpTypeAddress():
		if avp.Data.Size < 16 {
			binary.Write(buf, binary.BigEndian, uint16(1))
			binary.Write(buf, binary.BigEndian, (avp.Data.Value).(net.IP).To4())
		} else {
			binary.Write(buf, binary.BigEndian, uint16(2))
			binary.Write(buf, binary.BigEndian, (avp.Data.Value).(net.IP).To16())
		}
	case Dict.AvpTypeTime():
		binary.Write(buf, binary.BigEndian, (avp.Data.Value).(uint32))
	case Dict.AvpTypeUTF8String():
		buf.WriteString((avp.Data.Value).(string))
	case Dict.AvpTypeIdentity():
		buf.WriteString((avp.Data.Value).(string))
	case Dict.AvpTypeURI():
		buf.WriteString((avp.Data.Value).(string))
	case Dict.AvpTypeIPFilterRule():
		buf.WriteString((avp.Data.Value).(string))
	case Dict.AvpTypeQoSFilterRule():
		buf.WriteString((avp.Data.Value).(string))
	case Dict.AvpTypeEnumerated():
		binary.Write(buf, binary.BigEndian, (avp.Data.Value).(int32))
	case Dict.AvpTypeGrouped():
		for _, member := range avp.Data.Value.([]*Avp) {
			if err := member.Serialize(buf); err != nil {
				return err
			}
		}
	default:
		return &ErrUnknownAvpType{avp}
	}

	pad := alignTo4(uint32(buf.Len())) - uint32(buf.Len())
	for range pad {
		buf.WriteByte(0)
	}

	return nil
}

func (avp *Avp) Deserialize(data []byte) (uint32, error) {
	avpCode := binary.BigEndian.Uint32(data[:4])
	offset := 4
	avpFlags := uint8(data[offset])
	avpLength := binary.BigEndian.Uint32(data[offset:offset+4]) & 0x00FFFFFF
	offset += 4

	avpTmpl, err := Dict.GetAvpByCode(avpCode)
	if err != nil {
		avp.Code = avpCode
		return alignTo4(avpLength), err
	}
	*avp = *avpTmpl.clone()
	avp.Flags = avpFlags
	avp.Length = avpLength
	if avp.IsVendorSpec() {
		avp.VndId = binary.BigEndian.Uint32(data[offset : offset+4])
		offset += 4
	}
	data = data[offset:]
	dataSize := avp.Length - uint32(offset)

	avp.Data = new(AvpData)

	switch avp.Type {
	case Dict.AvpTypeOctetString():
		avp.Data.Value = data[:dataSize]
		avp.Data.Size = dataSize
	case Dict.AvpTypeInteger32():
		avp.Data.Value = int32(binary.BigEndian.Uint32(data[:4]))
		avp.Data.Size = 4
	case Dict.AvpTypeInteger64():
		avp.Data.Value = int64(binary.BigEndian.Uint64(data[:8]))
		avp.Data.Size = 8
	case Dict.AvpTypeUnsigned32():
		avp.Data.Value = binary.BigEndian.Uint32(data[:4])
		avp.Data.Size = 4
	case Dict.AvpTypeUnsigned64():
		avp.Data.Value = binary.BigEndian.Uint64(data[:8])
		avp.Data.Size = 8
	case Dict.AvpTypeFloat32():
		avp.Data.Value = math.Float32frombits(binary.BigEndian.Uint32(data[:4]))
		avp.Data.Size = 4
	case Dict.AvpTypeFloat64():
		avp.Data.Value = math.Float64frombits(binary.BigEndian.Uint64(data[:8]))
		avp.Data.Size = 8
	case Dict.AvpTypeAddress():
		addrType := binary.BigEndian.Uint16(data[:2])
		switch addrType {
		case 1: // IPv4
			avp.Data.Value, _ = netip.AddrFromSlice(data[2:6])
			avp.Data.Size = 4
		case 2: // IPv6
			avp.Data.Value, _ = netip.AddrFromSlice(data[2:18])
			avp.Data.Size = 16
		}
	case Dict.AvpTypeTime():
		avp.Data.Value = binary.BigEndian.Uint32(data[:4])
		avp.Data.Size = 4
	case Dict.AvpTypeUTF8String():
		avp.Data.Value = string(data[:dataSize])
		avp.Data.Size = dataSize
	case Dict.AvpTypeIdentity():
		avp.Data.Value = string(data[:dataSize])
		avp.Data.Size = dataSize
	case Dict.AvpTypeURI():
		avp.Data.Value = string(data[:dataSize])
		avp.Data.Size = dataSize
	case Dict.AvpTypeIPFilterRule():
		avp.Data.Value = string(data[:dataSize])
		avp.Data.Size = dataSize
	case Dict.AvpTypeQoSFilterRule():
		avp.Data.Value = string(data[:dataSize])
		avp.Data.Size = dataSize
	case Dict.AvpTypeEnumerated():
		avp.Data.Value = int32(binary.BigEndian.Uint32(data[:4]))
		avp.Data.Size = 4
	case Dict.AvpTypeGrouped():
		avp.Data.Value = make([]*Avp, 0)
		offset := uint32(0)
		for offset < dataSize {
			member := new(Avp)
			o, err := member.Deserialize(data[offset:])
			if err != nil {
				slog.Warn(err.Error())
			}

			offset += o
			avp.Data.Value = append(avp.Data.Value.([]*Avp), member)
		}
		avp.Data.Size = dataSize
	}

	return alignTo4(avp.Length), nil
}

func (avp *Avp) IsVendorSpec() bool {
	return (avp.Flags&Dict.AvpFlagV() != 0)
}

func (avp *Avp) IsMandatory() bool {
	return (avp.Flags&Dict.AvpFlagM() != 0)
}

func (avp *Avp) IsProtected() bool {
	return (avp.Flags&Dict.AvpFlagP() != 0)
}

func (avp *Avp) IsGrouped() bool {
	return avp.Type == Dict.AvpTypeGrouped()
}

func (avp *Avp) Dump(shift ...int) {
	tab := func(n int) {
		for range n {
			fmt.Print(" ")
		}
	}

	tab(shift[0])
	if len(avp.Name) > 0 {
		fmt.Printf("%s (%d): ", avp.Name, avp.Code)
	} else {
		fmt.Printf("Unknown <%d>", avp.Code)
	}

	if avp.IsGrouped() {
		fmt.Println()
		for _, member := range avp.Data.Value.([]*Avp) {
			member.Dump(shift[0] + 2)
		}
		return
	}

	if avp.Data != nil && avp.Data.Value != nil {
		if avp.Type == Dict.AvpTypeOctetString() {
			fmt.Printf("%x", avp.Data.Value.([]byte))
		} else {
			fmt.Printf("%v", avp.Data.Value)
		}
	}
	fmt.Println()
}

func (avp *Avp) clone() *Avp {
	if avp == nil {
		return nil
	}

	cloned := new(Avp)
	cloned.Code = avp.Code
	cloned.Name = avp.Name
	cloned.Length = avp.Length
	cloned.Flags = avp.Flags
	cloned.VndId = avp.VndId
	cloned.Type = avp.Type
	cloned.Enum = avp.Enum
	if avp.IsGrouped() {
		cloned.Group = new(Group)
		cloned.Group.Members = make([]*AvpRule, len(avp.Group.Members))
		for i, member := range avp.Group.Members {
			cloned.Group.Members[i] = new(AvpRule)
			*cloned.Group.Members[i] = *member
		}
	}
	return cloned
}

// -- Helpers
// --
func alignTo4(n uint32) uint32 {
	return ((n + 3) &^ 3)
}
