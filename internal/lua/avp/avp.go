//
// Project: TGDP - Traffic Generator for Diameter Protocol
// Description: Simple tool for testing and debugging the Diameter protocol
//
// Author: Alexander Kefeli <alexander.kefeli@gmail.com>
//
// File: avp.go
// Description: Lua API: AVP handling
//

package l_avp

import (
	"net/netip"
	"strings"
	"tgdp/internal/lua/l2g"
	"tgdp/pkg/diameter"

	lvm "github.com/yuin/gopher-lua"
)

// -- Consts
// --
const LuaTypeName = "avp"

// -- Variables
// --
var (
	metaTable *lvm.LTable
	methods   map[string]lvm.LGFunction
)

// -- Functions
// --
func New(L *lvm.LState) int {
	avpName := L.CheckString(1)
	avpCode := L.CheckInt(2)
	avpFlags := L.CheckInt(3)
	avpVendor_id := L.CheckInt(4)
	avpType := L.CheckInt(5)

	avp := diameter.NewAvp(avpName, uint32(avpCode), uint8(avpFlags), uint32(avpVendor_id), int(avpType))
	ud := L.NewUserData()
	ud.Value = avp
	L.SetMetatable(ud, metaTable)
	L.Push(ud)

	return 1
}

func Get(L *lvm.LState) int {
	avpId := l2g.CheckId(L, 1)
	if avp, err := diameter.Dict.GetAvp(avpId); err == nil {
		ud := L.NewUserData()
		ud.Value = avp
		L.SetMetatable(ud, metaTable)
		L.Push(ud)
		L.Push(lvm.LNil)
	} else {
		L.Push(lvm.LNil)
		L.Push(lvm.LString(err.Error()))
	}

	return 2
}

func GetValue(L *lvm.LState) int {
	if avp := Check(L, 1); avp != nil {
		return PushValue(L, avp)
	}

	return 0
}

func GetValues(L *lvm.LState) int {
	return 0
}

func PushValue(L *lvm.LState, avp *diameter.Avp) int {
	switch avp.Type {
	case diameter.Dict.AvpTypeOctetString():
		L.Push(lvm.LString(string(avp.Data.Value.([]byte))))
	case diameter.Dict.AvpTypeInteger32():
		L.Push(lvm.LNumber(avp.Data.Value.(int32)))
	case diameter.Dict.AvpTypeInteger64():
		L.Push(lvm.LNumber(avp.Data.Value.(int64)))
	case diameter.Dict.AvpTypeUnsigned32():
		L.Push(lvm.LNumber(avp.Data.Value.(uint32)))
	case diameter.Dict.AvpTypeUnsigned64():
		L.Push(lvm.LNumber(avp.Data.Value.(uint64)))
	case diameter.Dict.AvpTypeFloat32():
		L.Push(lvm.LNumber(avp.Data.Value.(float32)))
	case diameter.Dict.AvpTypeFloat64():
		L.Push(lvm.LNumber(avp.Data.Value.(float64)))
	case diameter.Dict.AvpTypeAddress():
		L.Push(lvm.LString(avp.Data.Value.(netip.Addr).StringExpanded()))
	case diameter.Dict.AvpTypeTime():
		L.Push(lvm.LNumber(avp.Data.Value.(uint32)))
	case diameter.Dict.AvpTypeUTF8String():
		L.Push(lvm.LString(avp.Data.Value.(string)))
	case diameter.Dict.AvpTypeIdentity():
		L.Push(lvm.LString(avp.Data.Value.(string)))
	case diameter.Dict.AvpTypeURI():
		L.Push(lvm.LString(avp.Data.Value.(string)))
	case diameter.Dict.AvpTypeIPFilterRule():
		L.Push(lvm.LString(avp.Data.Value.(string)))
	case diameter.Dict.AvpTypeQoSFilterRule():
		L.Push(lvm.LString(avp.Data.Value.(string)))
	case diameter.Dict.AvpTypeEnumerated():
		L.Push(lvm.LNumber(avp.Data.Value.(int32)))
	case diameter.Dict.AvpTypeGrouped():
		L.Push(lvm.LNil)
	default:
		L.Push(lvm.LNil)
		L.Push(lvm.LString((&diameter.ErrUnknownAvpType{Avp: avp}).Error()))
		return 2
	}
	L.Push(lvm.LNil)

	return 2
}

func SetValue(L *lvm.LState) int {
	if avp := Check(L, 1); avp != nil {
		if value := PopValue(L, 2, avp); value != nil {
			if err := avp.SetValue(&value); err == nil {
				return 0
			} else {
				L.Push(lvm.LString(err.Error()))
			}
		} else {
			L.Push(lvm.LString((&diameter.ErrInvalidAvpValue{Avp: avp, Value: value}).Error()))
		}
	}

	return 1
}

func PopValue(L *lvm.LState, n int, avp *diameter.Avp) any {
	value := L.CheckAny(n)
	switch avp.Type {
	case diameter.Dict.AvpTypeOctetString():
		return l2g.String(value)
	case diameter.Dict.AvpTypeInteger32():
		return l2g.Int32(value)
	case diameter.Dict.AvpTypeInteger64():
		return l2g.Int64(value)
	case diameter.Dict.AvpTypeUnsigned32():
		return l2g.UInt32(value)
	case diameter.Dict.AvpTypeUnsigned64():
		return l2g.UInt64(value)
	case diameter.Dict.AvpTypeFloat32():
		return l2g.Float32(value)
	case diameter.Dict.AvpTypeFloat64():
		return l2g.Float64(value)
	case diameter.Dict.AvpTypeAddress():
		return l2g.String(value)
	case diameter.Dict.AvpTypeTime():
		return l2g.String(value)
	case diameter.Dict.AvpTypeUTF8String():
		return l2g.String(value)
	case diameter.Dict.AvpTypeIdentity():
		return l2g.String(value)
	case diameter.Dict.AvpTypeURI():
		return l2g.String(value)
	case diameter.Dict.AvpTypeIPFilterRule():
		return l2g.String(value)
	case diameter.Dict.AvpTypeQoSFilterRule():
		return l2g.String(value)
	case diameter.Dict.AvpTypeEnumerated():
		switch value.Type() {
		case lvm.LTNumber:
			return l2g.Int(value)
		case lvm.LTString:
			return l2g.String(value)
		}
	case diameter.Dict.AvpTypeGrouped():
		// TODO: implement grouped AVP
		return nil
	}

	return nil
}

func IsGrouped(L *lvm.LState) int {
	if avp := Check(L, 1); avp != nil {
		L.Push(lvm.LBool(avp.IsGrouped()))
		return 1
	}
	return 0
}

func IsMandatory(L *lvm.LState) int {
	if avp := Check(L, 1); avp != nil {
		L.Push(lvm.LBool(avp.IsMandatory()))
		return 1
	}
	return 0
}

func MetaTable() *lvm.LTable {
	return metaTable
}

func Check(L *lvm.LState, n int) *diameter.Avp {
	ud := L.CheckUserData(n)
	if v, ok := ud.Value.(*diameter.Avp); ok {
		return v
	} else {
		L.ArgError(1, "AVP type expected")
		return nil
	}
}

func Register(L *lvm.LState) *lvm.LTable {
	metaTable = L.NewTypeMetatable(LuaTypeName)

	L.SetField(metaTable, "new", L.NewFunction(New))
	L.SetField(metaTable, "get", L.NewFunction(Get))
	L.SetField(metaTable, "__index", L.NewFunction(index))
	L.SetField(metaTable, "__newindex", L.NewFunction(newIndex))

	return metaTable
}

func index(L *lvm.LState) int {
	avp := Check(L, 1)
	key := L.CheckString(2)

	switch strings.ToLower(key) {
	case "code":
		L.Push(lvm.LNumber(avp.Code))
	case "flags":
		L.Push(lvm.LNumber(avp.Flags))
	case "name":
		L.Push(lvm.LString(avp.Name))
	case "vendor_id":
		L.Push(lvm.LNumber(avp.VndId))
	case "value":
		return PushValue(L, avp)
	case "members":
		if !avp.IsGrouped() {
			return 0
		}
		members := L.NewTable()
		for _, member := range avp.Data.Value.([]*diameter.Avp) {
			ud := L.NewUserData()
			ud.Value = member
			L.SetMetatable(ud, MetaTable())
			members.Append(ud)
		}
		L.Push(members)
	default:
		if fn, exists := methods[key]; exists {
			L.Push(L.NewFunction(fn))
		} else {
			L.Push(lvm.LNil)
		}
	}

	return 1
}

func newIndex(L *lvm.LState) int {
	if avp := Check(L, 1); avp != nil {
		key := L.CheckString(2)
		value := L.CheckAny(3)

		switch strings.ToLower(key) {
		case "code":
			avp.Code = l2g.UInt32(value)
		case "name":
			avp.Name = l2g.String(value)
		case "flags":
			avp.Flags = l2g.UInt8(value)
		case "vendor_id":
			avp.VndId = l2g.UInt32(value)
			avp.Flags |= diameter.Dict.AvpFlagV()
		case "value":
			v := PopValue(L, 3, avp)
			if err := avp.SetValue(&v); err != nil {
				L.Push(lvm.LString(err.Error()))
				return 1
			}
		}
	}

	return 0
}

func init() {
	methods = make(map[string]lvm.LGFunction)
	methods["get_value"] = GetValue
	methods["set_value"] = SetValue
	methods["is_grouped"] = IsGrouped
	methods["is_mandatory"] = IsMandatory
}
