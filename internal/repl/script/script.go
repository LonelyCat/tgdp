//
// Project: TGDP - Traffic Generator for Diameter Protocol
// Description: Simple tool for testing and debugging the Diameter protocol
//
// Author: Alexander Kefeli <alexander.kefeli@gmail.com>
//
// File: script.go
// Description: REPL: 'run' command implementation
//

package script

import (
	"fmt"

	"tgdp/internal/lua"

	"github.com/spf13/cobra"
	"github.com/chzyer/readline"
)

// -- Variables
// --
var (
	RootCommand = &cobra.Command{
		Use:     "run",
		Short:   "run <script.lua>",
		Long:    "Run a Lua script",
		Example: "run notify.lua msisdn 12345",
		Run:     script,
	}
)

// -- Functions
// --
func CompList() []readline.PrefixCompleterInterface {
	return []readline.PrefixCompleterInterface{readline.PcItem(
		RootCommand.Use,
	)}
}

func script(cmd *cobra.Command, args []string) {
	if len(args) < 1 {
		fmt.Println(cmd.Short)
		return
	}

	lua.Run(args)
}
