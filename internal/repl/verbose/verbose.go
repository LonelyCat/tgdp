//
// Project: TGDP - Traffic Generator for Diameter Protocol
// Description: Simple tool for testing and debugging the Diameter protocol
//
// Author: Alexander Kefeli <alexander.kefeli@gmail.com>
//
// File: verbose.go
// Description: REPL: 'verbose' command implementation
//

package verbose

import (
	"fmt"
	"strconv"

	"tgdp/pkg/diameter"

	"github.com/chzyer/readline"
	"github.com/spf13/cobra"
)

// -- Variables
// --
var (
	RootCommand = &cobra.Command{
		Use:                "verbose",
		Short:              "verbose [level]",
		Long:               "Set verbosity level",
		Example:            "verbose 3",
		Run:                verbose,
		DisableFlagParsing: true,
	}
)

// -- Functions
// --
func CompList() []readline.PrefixCompleterInterface {
	return []readline.PrefixCompleterInterface{readline.PcItem(
		RootCommand.Use,
	)}
}

func verbose(cmd *cobra.Command, args []string) {
	if len(args) == 0 {
		fmt.Printf("Verbosity level: %d\n", diameter.GetVerboseLevel())
		return
	}

	level, err := strconv.Atoi(args[0])
	if err != nil {
		fmt.Printf("Invalid verbosity level: %v", args[0])
		return
	}

	diameter.SetVerboseLevel(level)
}
