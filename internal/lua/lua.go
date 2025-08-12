//
// Project: TGDP - Traffic Generator for Diameter Protocol
// Description: Simple tool for testing and debugging the Diameter protocol
//
// Author: Alexander Kefeli <alexander.kefeli@gmail.com>
//
// File: lua.go
// Description: Lua API: execute a script
//

package lua

import (
	"log/slog"
	"strings"

	l_avp "tgdp/internal/lua/avp"
	l_msg "tgdp/internal/lua/message"
	l_peer "tgdp/internal/lua/peer"

	"tgdp/pkg/diameter"

	lvm "github.com/yuin/gopher-lua"
)

// -- Consts
// --
const moduleName = "diameter"

// -- Functions
// --
func Run(argv []string) {
	script := strings.TrimLeft(argv[0], "@")

	L := lvm.NewState()
	defer L.Close() //nolint:errcheck

	args := L.NewTable()
	L.SetTable(args, lvm.LNumber(0), lvm.LString(script))
	for i, arg := range argv[1:] {
		L.SetTable(args, lvm.LNumber(i+1), lvm.LString(arg))
	}
	L.SetGlobal("arg", args)

	L.PreloadModule(moduleName, modLoader)

	if err := L.DoFile(script); err != nil {
		slog.Error("Error executing script %s: %v", script, err)
	}
}

func modLoader(L *lvm.LState) int {
	module := L.NewTable()
	registerConstants(L, module)
	registerTypes(L, module)

	L.Push(module)
	return 1
}

func registerConstants(L *lvm.LState, module *lvm.LTable) {
	L.SetField(module, "MSG_FLAG_REQUEST", lvm.LNumber(diameter.Dict.CmdFlagR()))
	L.SetField(module, "MSG_FLAG_PROXYABLE", lvm.LNumber(diameter.Dict.CmdFlagP()))
	L.SetField(module, "MSG_FLAG_ERROR", lvm.LNumber(diameter.Dict.CmdFlagE()))
	L.SetField(module, "MSG_FLAG_RETRANSMISSION", lvm.LNumber(diameter.Dict.CmdFlagT()))

	L.SetField(module, "AVP_FLAG_VENDOR_SPECIFIC", lvm.LNumber(diameter.Dict.AvpFlagV()))
	L.SetField(module, "AVP_FLAG_MANDATORY", lvm.LNumber(diameter.Dict.AvpFlagM()))
	L.SetField(module, "AVP_FLAG_PROTECTED", lvm.LNumber(diameter.Dict.AvpFlagP()))

	L.SetField(module, "AVP_TYPE_OCTET_STRING", lvm.LNumber(diameter.Dict.AvpTypeOctetString()))
	L.SetField(module, "AVP_TYPE_INTEGER32", lvm.LNumber(diameter.Dict.AvpTypeInteger32()))
	L.SetField(module, "AVP_TYPE_INTEGER64", lvm.LNumber(diameter.Dict.AvpTypeInteger64()))
	L.SetField(module, "AVP_TYPE_UNSIGNED32", lvm.LNumber(diameter.Dict.AvpTypeUnsigned32()))
	L.SetField(module, "AVP_TYPE_UNSIGNED64", lvm.LNumber(diameter.Dict.AvpTypeUnsigned64()))
	L.SetField(module, "AVP_TYPE_FLOAT32", lvm.LNumber(diameter.Dict.AvpTypeFloat32()))
	L.SetField(module, "AVP_TYPE_FLOAT64", lvm.LNumber(diameter.Dict.AvpTypeFloat64()))
	L.SetField(module, "AVP_TYPE_ADDRESS", lvm.LNumber(diameter.Dict.AvpTypeAddress()))
	L.SetField(module, "AVP_TYPE_TIME", lvm.LNumber(diameter.Dict.AvpTypeTime()))
	L.SetField(module, "AVP_TYPE_UTF8_STRING", lvm.LNumber(diameter.Dict.AvpTypeUTF8String()))
	L.SetField(module, "AVP_TYPE_IDENTITY", lvm.LNumber(diameter.Dict.AvpTypeIdentity()))
	L.SetField(module, "AVP_TYPE_IP_FILTER_RULE", lvm.LNumber(diameter.Dict.AvpTypeIPFilterRule()))
	L.SetField(module, "AVP_TYPE_QOS_FILTER_RULE", lvm.LNumber(diameter.Dict.AvpTypeQoSFilterRule()))
	L.SetField(module, "AVP_TYPE_ENUMERATED", lvm.LNumber(diameter.Dict.AvpTypeEnumerated()))
	L.SetField(module, "AVP_TYPE_GROUPED", lvm.LNumber(diameter.Dict.AvpTypeGrouped()))

	// TODO: add erorrs constants
}

func registerTypes(L *lvm.LState, module *lvm.LTable) {
	L.SetField(module, l_peer.LuaTypeName, l_peer.Register(L))
	L.SetField(module, l_msg.LuaTypeName, l_msg.Register(L))
	L.SetField(module, l_avp.LuaTypeName, l_avp.Register(L))

	L.SetField(module, "write_pcap", L.NewFunction(writePcap))
	L.SetField(module, "dump", L.NewFunction(dump))
}
