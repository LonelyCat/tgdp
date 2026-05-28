//
// Project: TGDP - Traffic Generator for Diameter Protocol
// Description: Simple tool for testing and debugging the Diameter protocol
//
// Author: Alexander Kefeli <alexander.kefeli@gmail.com>
//
// File: version.go
// Description: REPL: 'version' command implementation
//

package version

import (
	ver "tgdp/internal/version"

	"github.com/chzyer/readline"
	"github.com/spf13/cobra"
)

// Variables
//

var (
	RootCommand = &cobra.Command{
		Use:                "version",
		Short:              "version",
		Long:               "Show version information",
		Example:            "version",
		Run:                version,
		DisableFlagParsing: true,
	}
)

// Functions
//

func CompList() []readline.PrefixCompleterInterface {
	return []readline.PrefixCompleterInterface{readline.PcItem(
		RootCommand.Use,
	)}
}

func version(cmd *cobra.Command, args []string) {
	ver.Show()
}
