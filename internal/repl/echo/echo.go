//
// Project: TGDP - Traffic Generator for Diameter Protocol
// Description: Simple tool for testing and debugging the Diameter protocol
//
// Author: Alexander Kefeli <alexander.kefeli@gmail.com>
//
// File: echo.go
// Description: REPL: 'echo' command implementation
//

package echo

import (
	"fmt"
	"strings"

	"github.com/chzyer/readline"
	"github.com/spf13/cobra"
)

// -- Variables
// --
var (
	RootCommand = &cobra.Command{
		Use:                "echo",
		Short:              "echo <text>",
		Long:               "Print a text message",
		Run:                echo,
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

func echo(cmd *cobra.Command, args []string) {
	fmt.Println(strings.Join(args, " "))
}
