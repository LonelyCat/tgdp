//
// Project: TGDP - Traffic Generator for Diameter Protocol
// Description: Simple tool for testing and debugging the Diameter protocol
//
// Author: Alexander Kefeli <alexander.kefeli@gmail.com>
//
// File: pcap.go
// Description: Lua API: write PCAP file
//

package lua

import (
	l_msg "tgdp/internal/lua/message"
	l_peer "tgdp/internal/lua/peer"

	"tgdp/pkg/diameter"

	lvm "github.com/yuin/gopher-lua"
)

// Functions
//

func writePcap(L *lvm.LState) int {
	msg := l_msg.Check(L, 1)
	node := l_peer.Check(L, 2)
	dir := L.ToBool(3)

	buf, err := msg.Serialize()
	if err != nil {
		L.Push(lvm.LString(err.Error()))
		return 1
	}

	env := L.Context().Value(diameter.EnvContext).(*diameter.Diameter)

	if err := env.Pcap().Write(buf, node, dir); err != nil {
		L.Push(lvm.LString(err.Error()))
	} else {
		L.Push(lvm.LNil)
	}

	return 1
}
