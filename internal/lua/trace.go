//
// Project: TGDP - Traffic Generator for Diameter Protocol
// Description: Simple tool for testing and debugging the Diameter protocol
//
// Author: Alexander Kefeli <alexander.kefeli@gmail.com>
//
// File: trace.go
// Description: Lua API: trace an object
//

package lua

import (
	"tgdp/pkg/diameter"

	lvm "github.com/yuin/gopher-lua"
)

// Functions
//

func trace(L *lvm.LState) int {
	ud := L.CheckUserData(1)
	if obj, ok := ud.Value.(diameter.ITrace); ok {
		obj.Trace(0)
	} else {
		L.Push(lvm.LString("Unknown type"))
		return 1
	}
	return 0
}
