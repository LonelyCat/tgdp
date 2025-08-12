//
// Project: TGDP - Traffic Generator for Diameter Protocol
// Description: Simple tool for testing and debugging the Diameter protocol
//
// Author: Alexander Kefeli <alexander.kefeli@gmail.com>
//
// File: l2g.go
// Description: Lua API: Lua to Go datatypes conversation
//

package l2g

import (
	"tgdp/pkg/diameter"
	"tgdp/pkg/diameter/net/node"

	lvm "github.com/yuin/gopher-lua"
)

// -- Functions
// --
func CheckId(L *lvm.LState, n int) any {
	id := L.CheckAny(n)

	switch id.Type() {
	case lvm.LTNumber:
		return UInt32(id)
	case lvm.LTString:
		return String(id)
	case lvm.LTUserData:
		switch id.(*lvm.LUserData).Value.(type) {
		case *node.Node:
			return id.(*lvm.LUserData).Value.(*node.Node)
		case *diameter.Message:
			return id.(*lvm.LUserData).Value.(*diameter.Message)
		case *diameter.Avp:
			return id.(*lvm.LUserData).Value.(*diameter.Avp)
		}
	}

	return id
}

func Int(n any) int {
	return int(n.(lvm.LNumber))
}

func Int8(n any) int8 {
	return int8(n.(lvm.LNumber))
}

func Int16(n any) int16 {
	return int16(n.(lvm.LNumber))
}

func Int32(n any) int32 {
	return int32(n.(lvm.LNumber))
}

func Int64(n any) int64 {
	return int64(n.(lvm.LNumber))
}

func UInt8(n any) uint8 {
	return uint8(n.(lvm.LNumber))
}

func UInt16(n any) uint16 {
	return uint16(n.(lvm.LNumber))
}

func UInt32(n any) uint32 {
	return uint32(n.(lvm.LNumber))
}

func UInt64(n any) uint64 {
	return uint64(n.(lvm.LNumber))
}

func Float32(n any) float32 {
	return float32(n.(lvm.LNumber))
}

func Float64(n any) float64 {
	return float64(n.(lvm.LNumber))
}

func String(s any) string {
	return s.(lvm.LString).String()
}
