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
	"tgdp/pkg/diameter/diwe"

	lvm "github.com/yuin/gopher-lua"
)

// Consts
//

const (
	LuaTypeName = "avp"
)

const (
	keyCode     = "code"
	keyFlags    = "flags"
	keyName     = "name"
	keyVendorId = "vendor_id"
	keyValue    = "value"
	keyMembers  = "members"
)

// Variables
//

var (
	metaTable *lvm.LTable
	methods   map[string]lvm.LGFunction
)

// Functions
//

// New creates new AVP from parameters.
func New(L *lvm.LState) int {
	avpName := L.CheckString(1)
	avpCode := L.CheckInt(2)
	avpFlags := L.CheckInt(3)
	avpVendor_id := L.CheckInt(4)
	avpType := L.CheckInt(5)

	env := L.Context().Value(diameter.EnvContext).(*diameter.Diameter)
	avp := env.NewAvp(avpName, uint32(avpCode), uint8(avpFlags), uint32(avpVendor_id), int(avpType))
	ud := L.NewUserData()
	ud.Value = avp
	L.SetMetatable(ud, metaTable)
	L.Push(ud)

	return 1
}

// Fetch returns AVP instancies with values from the environment AvpStore.
// AVP can be defined by ID or name.
func Fetch(L *lvm.LState) int {
	avpId := l2g.CheckId(L, 1)
	env := L.Context().Value(diameter.EnvContext).(*diameter.Diameter)
	avpd, err := env.Dict().GetAvp(avpId)
	if err != nil {
		L.Push(lvm.LNil)
		L.Push(lvm.LString(err.Error()))
		return 2
	}

	avps := L.NewTable()
	for _, avp := range env.Store().Fetch(avpd.Code) {
		ud := L.NewUserData()
		ud.Value = avp
		L.SetMetatable(ud, MetaTable())
		avps.Append(ud)
	}

	L.Push(avps)
	L.Push(lvm.LNil)
	return 2
}

func GetValue(L *lvm.LState) int {
	if avp := Check(L, 1); avp != nil {
		return PushValue(L, avp)
	}

	return 0
}

func PushValue(L *lvm.LState, avp *diameter.Avp) int {
	env := L.Context().Value(diameter.EnvContext).(*diameter.Diameter)

	switch avp.Type() {
	case env.Dict().AvpDataType().OctetString:
		L.Push(lvm.LString(string(avp.Data().Value.([]byte))))
	case env.Dict().AvpDataType().Integer32:
		L.Push(lvm.LNumber(avp.Data().Value.(int32)))
	case env.Dict().AvpDataType().Integer64:
		L.Push(lvm.LNumber(avp.Data().Value.(int64)))
	case env.Dict().AvpDataType().Unsigned32:
		L.Push(lvm.LNumber(avp.Data().Value.(uint32)))
	case env.Dict().AvpDataType().Unsigned64:
		L.Push(lvm.LNumber(avp.Data().Value.(uint64)))
	case env.Dict().AvpDataType().Float32:
		L.Push(lvm.LNumber(avp.Data().Value.(float32)))
	case env.Dict().AvpDataType().Float64:
		L.Push(lvm.LNumber(avp.Data().Value.(float64)))
	case env.Dict().AvpDataType().Address:
		L.Push(lvm.LString(avp.Data().Value.(netip.Addr).StringExpanded()))
	case env.Dict().AvpDataType().Time:
		L.Push(lvm.LNumber(avp.Data().Value.(uint32)))
	case env.Dict().AvpDataType().UTF8String:
		L.Push(lvm.LString(avp.Data().Value.(string)))
	case env.Dict().AvpDataType().Identity:
		L.Push(lvm.LString(avp.Data().Value.(string)))
	case env.Dict().AvpDataType().URI:
		L.Push(lvm.LString(avp.Data().Value.(string)))
	case env.Dict().AvpDataType().IPFilterRule:
		L.Push(lvm.LString(avp.Data().Value.(string)))
	case env.Dict().AvpDataType().QoSFilterRule:
		L.Push(lvm.LString(avp.Data().Value.(string)))
	case env.Dict().AvpDataType().Enumerated:
		L.Push(lvm.LNumber(avp.Data().Value.(int32)))
	case env.Dict().AvpDataType().Grouped:
		members := L.NewTable()
		for _, member := range avp.Value().([]*diameter.Avp) {
			ud := L.NewUserData()
			ud.Value = member
			L.SetMetatable(ud, MetaTable())
			members.Append(ud)
		}
		L.Push(members)
	default:
		L.Push(lvm.LNil)
		L.Push(lvm.LString((&diwe.ErrUnknownAvpType{Avp: avp.Name()}).Error()))
		return 2
	}
	L.Push(lvm.LNil)

	return 2
}

func SetValue(L *lvm.LState) int {
	if avp := Check(L, 1); avp != nil {
		if value := PopValue(L, 2, avp); value != nil {
			if err := avp.SetValue(value); err == nil {
				return 0
			} else {
				L.Push(lvm.LString(err.Error()))
			}
		} else {
			L.Push(lvm.LString((&diwe.ErrInvalidAvpValue{Avp: avp.Name(), Value: value}).Error()))
		}
	}

	return 1
}

func PopValue(L *lvm.LState, n int, avp *diameter.Avp) any {
	env := L.Context().Value(diameter.EnvContext).(*diameter.Diameter)

	value := L.CheckAny(n)

	switch avp.Type() {
	case env.Dict().AvpDataType().OctetString:
		return l2g.String(value)
	case env.Dict().AvpDataType().Integer32:
		return l2g.Int32(value)
	case env.Dict().AvpDataType().Integer64:
		return l2g.Int64(value)
	case env.Dict().AvpDataType().Unsigned32:
		return l2g.UInt32(value)
	case env.Dict().AvpDataType().Unsigned64:
		return l2g.UInt64(value)
	case env.Dict().AvpDataType().Float32:
		return l2g.Float32(value)
	case env.Dict().AvpDataType().Float64:
		return l2g.Float64(value)
	case env.Dict().AvpDataType().Address:
		return l2g.String(value)
	case env.Dict().AvpDataType().Time:
		return l2g.String(value)
	case env.Dict().AvpDataType().UTF8String:
		return l2g.String(value)
	case env.Dict().AvpDataType().Identity:
		return l2g.String(value)
	case env.Dict().AvpDataType().URI:
		return l2g.String(value)
	case env.Dict().AvpDataType().IPFilterRule:
		return l2g.String(value)
	case env.Dict().AvpDataType().QoSFilterRule:
		return l2g.String(value)
	case env.Dict().AvpDataType().Enumerated:
		switch value.Type() {
		case lvm.LTNumber:
			return l2g.Int(value)
		case lvm.LTString:
			return l2g.String(value)
		}
	case env.Dict().AvpDataType().Grouped:
		if t, ok := value.(*lvm.LTable); ok {
			members := make([]*diameter.Avp, 0, t.Len())
			t.ForEach(func(k, v lvm.LValue) {
				if member, ok := v.(*lvm.LUserData); ok {
					members = append(members, member.Value.(*diameter.Avp))
				}
			})
			return members
		}
	}

	return nil
}

// Totext
func ToText(L *lvm.LState) int {
	avp := Check(L, 1)
	if avp == nil {
		return 0
	}

	env := L.Context().Value(diameter.EnvContext).(*diameter.Diameter)

	if codec, exists := env.Codec(avp.Type()); exists {
		L.Push(lvm.LString(codec.ToText(avp)))
		return 1
	}

	return 0
}

// IsMandatory returns true if AVP is mandatory to present in message.
func IsMandatory(L *lvm.LState) int {
	if avp := Check(L, 1); avp != nil {
		L.Push(lvm.LBool(avp.IsMandatory()))
		return 1
	}

	return 0
}

// Gouped type AVP specific
//

// IsGrouped returns true if the AVP is a grouped type.
func IsGrouped(L *lvm.LState) int {
	if avp := Check(L, 1); avp != nil {
		L.Push(lvm.LBool(avp.IsGrouped()))
		return 1
	}

	return 0
}

// Members returns list of members for the grouped AVP.
func Members(L *lvm.LState) int {
	avp := Check(L, 1)
	if avp == nil {
		return 0
	}

	env := L.Context().Value(diameter.EnvContext).(*diameter.Diameter)

	avpD, err := env.Dict().GetAvp(avp.Code())
	if err != nil {
		return 0
	}

	membersList := L.NewTable()
	for i, member := range avpD.Group.Members {
		membersList.RawSetInt(i+1, lvm.LString(member.Name))
	}
	L.Push(membersList)

	return 1
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
	L.SetField(metaTable, "fetch", L.NewFunction(Fetch))
	L.SetField(metaTable, "__index", L.NewFunction(index))
	L.SetField(metaTable, "__newindex", L.NewFunction(newIndex))

	return metaTable
}

func index(L *lvm.LState) int {
	avp := Check(L, 1)
	key := L.CheckString(2)

	switch strings.ToLower(key) {
	case keyCode:
		L.Push(lvm.LNumber(avp.Code()))
	case keyFlags:
		L.Push(lvm.LNumber(avp.Flags()))
	case keyName:
		L.Push(lvm.LString(avp.Name()))
	case keyVendorId:
		L.Push(lvm.LNumber(avp.VendorId()))
	case keyValue:
		return PushValue(L, avp)
	case keyMembers:
		return Members(L)
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
	avp := Check(L, 1)
	if avp == nil {
		return 0
	}

	key := L.CheckString(2)
	// value := L.CheckAny(3)

	switch strings.ToLower(key) {
	// case "code":
	// 	avp.Code = l2g.UInt32(value)
	// case "name":
	// 	avp.Name = l2g.String(value)
	// case "flags":
	// 	avp.Flags = l2g.UInt8(value)
	// case "vendor_id":
	// 	avp.VndId = l2g.UInt32(value)
	// 	avp.Flags |= diameter.Dict.AvpFlagV()
	case keyValue:
		v := PopValue(L, 3, avp)
		if err := avp.SetValue(v); err != nil {
			L.Push(lvm.LString(err.Error()))
			return 1
		}
	}

	return 0
}

// Init
//

func init() {
	methods = make(map[string]lvm.LGFunction)
	methods["get_value"] = GetValue
	methods["set_value"] = SetValue
	methods["to_text"] = ToText
	methods["is_grouped"] = IsGrouped
	methods["is_mandatory"] = IsMandatory
}
