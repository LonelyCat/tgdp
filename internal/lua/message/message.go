//
// Project: TGDP - Traffic Generator for Diameter Protocol
// Description: Simple tool for testing and debugging the Diameter protocol
//
// Author: Alexander Kefeli <alexander.kefeli@gmail.com>
//
// File: message.go
// Description: Lua API: message handling
//

package l_msg

import (
	"strings"
	l_avp "tgdp/internal/lua/avp"
	"tgdp/internal/lua/l2g"
	"tgdp/pkg/diameter"

	lvm "github.com/yuin/gopher-lua"
)

// Constants
//

const (
	LuaTypeName = "message"
)

const (
	keyAppId    = "app_id"
	keyAppName  = "app_name"
	keyCmdCode  = "cmd_code"
	keyFlags    = "flags"
	keyEndToEnd = "end_to_end"
	keyHopByHop = "hop_by_hop"
	keyAvps     = "avps"
)

// Variables
//

var (
	metaTable *lvm.LTable
	methods   map[string]lvm.LGFunction
)

// Functions
//

func New(L *lvm.LState) int {
	return newMessage(L, false)
}

func Fetch(L *lvm.LState) int {
	return newMessage(L, true)
}

func newMessage(L *lvm.LState, fecthAvp bool) int {
	appId := l2g.CheckId(L, 1)
	cmdId := l2g.CheckId(L, 2)
	request := L.CheckBool(3)

	env := L.Context().Value(diameter.EnvContext).(*diameter.Diameter)

	app, err := env.Dict().GetApp(appId)
	if err != nil {
		L.Push(lvm.LNil)
		L.Push(lvm.LString(err.Error()))
		return 2
	}

	cmd, err := env.Dict().GetCmd(cmdId, app)
	if err != nil {
		L.Push(lvm.LNil)
		L.Push(lvm.LString(err.Error()))
		return 2
	}

	msg, err := env.NewMessage(app, cmd, request, fecthAvp)
	if err != nil {
		L.Push(lvm.LNil)
		L.Push(lvm.LString(err.Error()))
	} else {
		ud := L.NewUserData()
		ud.Value = msg
		L.SetMetatable(ud, metaTable)
		L.Push(ud)
		L.Push(lvm.LNil)
	}
	return 2
}

func AddAvp(L *lvm.LState) int {
	if msg := Check(L, 1); msg != nil {
		err := msg.AddAvp(l2g.CheckId(L, 2))
		if err != nil {
			L.Push(lvm.LString(err.Error()))
			return 1
		}
	}
	return 0
}

func GetAvp(L *lvm.LState) int {
	if msg := Check(L, 1); msg != nil {
		avp, err := msg.GetAvp(l2g.CheckId(L, 2))
		if err != nil {
			L.Push(lvm.LNil)
			L.Push(lvm.LString(err.Error()))
		} else {
			ud := L.NewUserData()
			ud.Value = avp
			L.SetMetatable(ud, l_avp.MetaTable())
			L.Push(ud)
			L.Push(lvm.LNil)
		}
	}
	return 2
}

func RemoveAvp(L *lvm.LState) int {
	if msg := Check(L, 1); msg != nil {
		err := msg.RemoveAvp(l2g.CheckId(L, 2))
		if err != nil {
			L.Push(lvm.LString(err.Error()))
			return 1
		}
	}
	return 0
}

func GetAvpValue(L *lvm.LState) int {
	if msg := Check(L, 1); msg != nil {
		avp, err := msg.GetAvp(l2g.CheckId(L, 2))
		if err != nil {
			L.Push(lvm.LString(err.Error()))
			return 1
		}
		return l_avp.PushValue(L, avp)
	}
	return 0
}

func SetAvpValue(L *lvm.LState) int {
	if msg := Check(L, 1); msg != nil {
		avp, err := msg.GetAvp(l2g.CheckId(L, 2))
		if err != nil {
			L.Push(lvm.LString(err.Error()))
			return 1
		}
		value := l_avp.PopValue(L, 3, avp)
		err = avp.SetValue(&value)
		if err != nil {
			L.Push(lvm.LString(err.Error()))
			return 1
		}
	}
	return 0
}

func IsRequest(L *lvm.LState) int {
	if msg := Check(L, 1); msg != nil {
		L.Push(lvm.LBool(msg.IsRequest()))
		return 1
	}
	return 0
}

func MetaTable() *lvm.LTable {
	return metaTable
}

func Check(L *lvm.LState, n int) *diameter.Message {
	ud := L.CheckUserData(n)
	if v, ok := ud.Value.(*diameter.Message); ok {
		return v
	} else {
		L.ArgError(1, "Message type expected")
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
	msg := Check(L, 1)
	key := L.CheckString(2)
	env := L.Context().Value(diameter.EnvContext).(*diameter.Diameter)

	switch strings.ToLower(key) {
	case keyAppId:
		L.Push(lvm.LNumber(msg.AppId))
	case keyAppName:
		app, err := env.Dict().GetApp(msg.AppId)
		if err != nil {
			L.Push(lvm.LString(err.Error()))
		} else {
			L.Push(lvm.LString(app.Name))
		}
	case keyCmdCode:
		L.Push(lvm.LNumber(msg.CmdCode))
	case keyFlags:
		L.Push(lvm.LNumber(msg.Flags))
	case keyHopByHop:
		L.Push(lvm.LNumber(msg.HopByHop))
	case keyEndToEnd:
		L.Push(lvm.LNumber(msg.EndToEnd))
	case keyAvps:
		avps := L.NewTable()
		for _, avp := range msg.Avps() {
			ud := L.NewUserData()
			ud.Value = avp
			L.SetMetatable(ud, l_avp.MetaTable())
			avps.Append(ud)
		}
		L.Push(avps)
	default:
		if fn, exists := methods[key]; exists {
			L.Push(L.NewFunction(fn))
		} else {
			avp, err := msg.GetAvp(l2g.CheckId(L, 2))
			if err == nil {
				return l_avp.PushValue(L, avp)
			}
			L.Push(lvm.LString(err.Error()))
		}
	}

	return 1
}

func newIndex(L *lvm.LState) int {
	msg := Check(L, 1)
	key := L.CheckString(2)
	value := L.CheckAny(3)

	switch strings.ToLower(key) {
	case keyAppId:
		msg.AppId = l2g.UInt32(value)
	case keyCmdCode:
		msg.CmdCode = l2g.UInt32(value)
	case keyFlags:
		msg.Flags = l2g.UInt8(value)
	case keyEndToEnd:
		msg.EndToEnd = l2g.UInt32(value)
	case keyHopByHop:
		msg.HopByHop = l2g.UInt32(value)
	}

	return 0
}

// Init
//

func init() {
	methods = make(map[string]lvm.LGFunction)
	methods["add_avp"] = AddAvp
	methods["get_avp"] = GetAvp
	methods["remove_avp"] = RemoveAvp
	methods["get_avp_value"] = GetAvpValue
	methods["set_avp_value"] = SetAvpValue
	methods["is_request"] = IsRequest
}
