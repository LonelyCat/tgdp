//
// Project: TGDP - Traffic Generator for Diameter Protocol
// Description: Simple tool for testing and debugging the Diameter protocol
//
// Author: Alexander Kefeli <alexander.kefeli@gmail.com>
//
// File: peer.go
// Description: Lua API: peer-to-peer network processing
//

package l_peer

import (
	"strings"

	l_msg "tgdp/internal/lua/message"
	"tgdp/pkg/diameter/net/node"

	lvm "github.com/yuin/gopher-lua"
)

// -- Consts
// --
const LuaTypeName = "peer"

// -- Variables
// --
var (
	metaTable *lvm.LTable
	methods   map[string]lvm.LGFunction
)

// -- Functions
// --
func New(L *lvm.LState) int {
	name := L.ToString(1)
	addr := L.ToString(2)
	port := L.ToInt(3)
	proto := L.ToString(4)
	timeout := L.ToInt(5)

	peer, err := node.New(name, addr, port, proto, timeout)
	if peer == nil {
		L.Push(lvm.LNil)
		L.Push(lvm.LString(err.Error()))
	} else {
		ud := L.NewUserData()
		ud.Value = peer
		L.SetMetatable(ud, metaTable)
		L.Push(ud)
		L.Push(lvm.LNil)
	}

	return 2
}

func Get(L *lvm.LState) int {
	name := L.ToString(1)
	peer, err := node.GetByName(name)
	if peer == nil {
		L.Push(lvm.LNil)
		L.Push(lvm.LString(err.Error()))
	} else {
		ud := L.NewUserData()
		ud.Value = peer
		L.SetMetatable(ud, metaTable)
		L.Push(ud)
		L.Push(lvm.LNil)
	}

	return 2
}

func Connect(L *lvm.LState) int {
	if peer := Check(L, 1); peer != nil {
		err := peer.Connect(true)
		if err != nil {
			L.Push(lvm.LString(err.Error()))
		} else {
			L.Push(lvm.LNil)
		}
	}

	return 1
}

func Disconnect(L *lvm.LState) int {
	if peer := Check(L, 1); peer != nil {
		err := peer.Disconnect(true, false)
		if err != nil {
			L.Push(lvm.LString(err.Error()))
		} else {
			L.Push(lvm.LNil)
		}
	}

	return 1
}

func SendTo(L *lvm.LState) int {
	if peer := Check(L, 1); peer != nil {
		if msg := l_msg.Check(L, 2); msg != nil {
			if err := peer.SendTo(msg); err != nil {
				L.Push(lvm.LString(err.Error()))
			} else {
				L.Push(lvm.LNil)
			}
		}

		return 1
	}

	return 0
}

func RecvFrom(L *lvm.LState) int {
	if peer := Check(L, 1); peer != nil {
		msg, err := peer.RecvFrom()
		if err != nil {
			L.Push(lvm.LNil)
			L.Push(lvm.LString(err.Error()))
		} else {
			ud := L.NewUserData()
			ud.Value = msg
			L.SetMetatable(ud, l_msg.MetaTable())
			L.Push(ud)
			L.Push(lvm.LNil)
		}

		return 2
	}

	return 0
}

func SetTimeout(L *lvm.LState) int {
	if peer := Check(L, 1); peer != nil {
		timeout := L.ToInt(2)
		if err := peer.SetTimeout(timeout); err != nil {
			L.Push(lvm.LString(err.Error()))
		}

	}

	return 0
}

func MetaTable() *lvm.LTable {
	return metaTable
}

func Check(L *lvm.LState, n int) *node.Node {
	ud := L.CheckUserData(n)
	if v, ok := ud.Value.(*node.Node); ok {
		return v
	} else {
		L.ArgError(1, "Peer type expected")
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
	peer := Check(L, 1)
	key := L.CheckString(2)

	if fn, exists := methods[key]; exists {
		L.Push(L.NewFunction(fn))
		return 1
	}

	switch strings.ToLower(key) {
	case node.KeyName:
		L.Push(lvm.LString(peer.Name))
	case node.KeyAddress:
		L.Push(lvm.LString(peer.Address))
	case node.KeyPort:
		L.Push(lvm.LNumber(peer.RemotePort))
	case node.KeyTransport:
		L.Push(lvm.LString(peer.Transport))
	case node.KeyTimeout:
		L.Push(lvm.LNumber(peer.Timeout))
	default:
		L.Push(lvm.LNil)
	}

	return 1
}

func newIndex(L *lvm.LState) int {
	return 0
}

func init() {
	methods = make(map[string]lvm.LGFunction)
	methods["connect"] = Connect
	methods["disconnect"] = Disconnect
	methods["send_to"] = SendTo
	methods["recv_from"] = RecvFrom
	methods["set_timeout"] = SetTimeout
}
