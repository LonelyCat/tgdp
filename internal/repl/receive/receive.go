//
// Project: TGDP - Traffic Generator for Diameter Protocol
// Description: Simple tool for testing and debugging the Diameter protocol
//
// Author: Alexander Kefeli <alexander.kefeli@gmail.com>
//
// File: receive.go
// Description: REPL: 'receive' command implementation
//

package receive

import (
	"fmt"

	"tgdp/internal/cli"
	"tgdp/internal/repl/comp"
	"tgdp/internal/repl/list"

	"github.com/chzyer/readline"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// -- Constants
// --
const flagWait = "wait"

// -- Variables
// --
var (
	RootCommand = &cobra.Command{
		Use:     "receive",
		Aliases: []string{"recv"},
		Short:   "receive [-w | --wait] <peer>",
		Long:    "Receive a message from a peer",
		Example: "receive -w HSS",
		Run:     receive,
	}
)

// -- Functions
// --
func CompList() []readline.PrefixCompleterInterface {
	pciList := []readline.PrefixCompleterInterface{}

	RootCommand.Flags().VisitAll(func(f *pflag.Flag) {
		pciList = append(pciList, readline.PcItem("-"+f.Shorthand, comp.PeerList(false)...))
		pciList = append(pciList, readline.PcItem("--"+f.Name, comp.PeerList(false)...))
	})

	pciList = append(pciList, comp.PeerList(false)...)

	return []readline.PrefixCompleterInterface{readline.PcItem(RootCommand.Use, pciList...)}
}

func init() {
	RootCommand.Flags().BoolP(flagWait, "w", false, "wait for data")
}

func receive(cmd *cobra.Command, args []string) {
	if len(args) != 1 {
		fmt.Println(cmd.Short)
		return
	}

	fw, _ := cmd.Flags().GetBool(flagWait)

	args[0] = list.PeerNameById(args[0])
	cli.Receive(args[0], fw) //nolint:errcheck

	f := cmd.Flags().Lookup(flagWait)
	f.Value.Set(f.DefValue) //nolint:errcheck
}
