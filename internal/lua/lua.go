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
	"context"
	"log/slog"
	"strings"
	"sync"

	l_avp "tgdp/internal/lua/avp"
	l_msg "tgdp/internal/lua/message"
	l_peer "tgdp/internal/lua/peer"

	"tgdp/pkg/diameter"

	lvm "github.com/yuin/gopher-lua"
)

// Consts
//

const moduleName = "diameter"

// Variables
//

var statePool = sync.Pool{
	New: func() any {
		L := lvm.NewState()
		// Preload types (they are static)
		module := L.NewTable()
		registerTypes(L, module)
		L.SetGlobal(moduleName, module)
		return L
	},
}

// Functions
//
// Run executes a Lua script with the given Diameter environment and arguments.
func Run(env *diameter.Diameter, argv []string) {
	script := strings.TrimLeft(argv[0], "@")

	L := statePool.Get().(*lvm.LState)
	defer statePool.Put(L)

	bg := context.Background()
	ctx := context.WithValue(bg, diameter.EnvContext, env)
	L.SetContext(ctx)

	// Update constants for current environment
	if module, ok := L.GetGlobal(moduleName).(*lvm.LTable); ok {
		registerConstants(L, module)
	}

	args := L.NewTable()
	L.SetTable(args, lvm.LNumber(0), lvm.LString(script))
	for i, arg := range argv[1:] {
		L.SetTable(args, lvm.LNumber(i+1), lvm.LString(arg))
	}
	L.SetGlobal("arg", args)

	if err := L.DoFile(script); err != nil {
		slog.Error("Error executing script", slog.String("script", script), slog.Any("error", err))
	}
}

func registerConstants(L *lvm.LState, module *lvm.LTable) {
	env := L.Context().Value(diameter.EnvContext).(*diameter.Diameter)

	L.SetField(module, "MSG_FLAG_REQUEST", lvm.LNumber(env.Dict().CmdFlag().R))
	L.SetField(module, "MSG_FLAG_PROXYABLE", lvm.LNumber(env.Dict().CmdFlag().P))
	L.SetField(module, "MSG_FLAG_ERROR", lvm.LNumber(env.Dict().CmdFlag().E))
	L.SetField(module, "MSG_FLAG_RETRANSMISSION", lvm.LNumber(env.Dict().CmdFlag().T))

	L.SetField(module, "AVP_FLAG_VENDOR_SPECIFIC", lvm.LNumber(env.Dict().AvpFlag().V))
	L.SetField(module, "AVP_FLAG_MANDATORY", lvm.LNumber(env.Dict().AvpFlag().M))
	L.SetField(module, "AVP_FLAG_PROTECTED", lvm.LNumber(env.Dict().AvpFlag().P))

	L.SetField(module, "AVP_TYPE_OCTET_STRING", lvm.LNumber(env.Dict().AvpDataType().OctetString))
	L.SetField(module, "AVP_TYPE_INTEGER32", lvm.LNumber(env.Dict().AvpDataType().Integer32))
	L.SetField(module, "AVP_TYPE_INTEGER64", lvm.LNumber(env.Dict().AvpDataType().Integer64))
	L.SetField(module, "AVP_TYPE_UNSIGNED32", lvm.LNumber(env.Dict().AvpDataType().Unsigned32))
	L.SetField(module, "AVP_TYPE_UNSIGNED64", lvm.LNumber(env.Dict().AvpDataType().Unsigned64))
	L.SetField(module, "AVP_TYPE_FLOAT32", lvm.LNumber(env.Dict().AvpDataType().Float32))
	L.SetField(module, "AVP_TYPE_FLOAT64", lvm.LNumber(env.Dict().AvpDataType().Float64))
	L.SetField(module, "AVP_TYPE_ADDRESS", lvm.LNumber(env.Dict().AvpDataType().Address))
	L.SetField(module, "AVP_TYPE_TIME", lvm.LNumber(env.Dict().AvpDataType().Time))
	L.SetField(module, "AVP_TYPE_UTF8_STRING", lvm.LNumber(env.Dict().AvpDataType().UTF8String))
	L.SetField(module, "AVP_TYPE_IDENTITY", lvm.LNumber(env.Dict().AvpDataType().Identity))
	L.SetField(module, "AVP_TYPE_IP_FILTER_RULE", lvm.LNumber(env.Dict().AvpDataType().IPFilterRule))
	L.SetField(module, "AVP_TYPE_QOS_FILTER_RULE", lvm.LNumber(env.Dict().AvpDataType().QoSFilterRule))
	L.SetField(module, "AVP_TYPE_ENUMERATED", lvm.LNumber(env.Dict().AvpDataType().Enumerated))
	L.SetField(module, "AVP_TYPE_GROUPED", lvm.LNumber(env.Dict().AvpDataType().Grouped))

	// TODO: add erorrs constants
}

func registerTypes(L *lvm.LState, module *lvm.LTable) {
	L.SetField(module, l_peer.LuaTypeName, l_peer.Register(L))
	L.SetField(module, l_msg.LuaTypeName, l_msg.Register(L))
	L.SetField(module, l_avp.LuaTypeName, l_avp.Register(L))

	L.SetField(module, "write_pcap", L.NewFunction(writePcap))
	L.SetField(module, "dump", L.NewFunction(trace))
}
