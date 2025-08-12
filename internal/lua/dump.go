//
// Project: TGDP - Traffic Generator for Diameter Protocol
// Description: Simple tool for testing and debugging the Diameter protocol
//
// Author: Alexander Kefeli <alexander.kefeli@gmail.com>
//
// File: dump.go
// Description: Lua API: dump an object
//

package lua

import (
	"tgdp/pkg/diameter"

	lvm "github.com/yuin/gopher-lua"
)

// -- Functions
// --
func dump(L *lvm.LState) int {
	ud := L.CheckUserData(1)
	if dbg, ok := ud.Value.(diameter.IDebug); ok {
		dbg.Dump(0)
	} else {
		L.Push(lvm.LString("Unknown type"))
		return 1
	}
	return 0
}
