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
	"tgdp/pkg/diameter/pcap"

	lvm "github.com/yuin/gopher-lua"
)

// -- Functions
// --
func writePcap(L *lvm.LState) int {
	msg := l_msg.Check(L, 1)
	node := l_peer.Check(L, 2)
	file := L.ToString(3)
	append := L.ToBool(4)

	buf, err := msg.Serialize()
	if err != nil {
		L.Push(lvm.LString(err.Error()))
		return 1
	}

	if err := pcap.Write(file, append, buf, node, msg.IsRequest()); err != nil {
		L.Push(lvm.LString(err.Error()))
	} else {
		L.Push(lvm.LNil)
	}

	return 1
}
