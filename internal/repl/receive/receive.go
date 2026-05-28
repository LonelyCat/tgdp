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
	"tgdp/internal/repl/peer"
	"tgdp/pkg/diameter"
	"tgdp/pkg/diameter/diwe"

	"github.com/chzyer/readline"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// Variables
//

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

var (
	flagWait bool
)

// Functions
//

func CompList(env *diameter.Diameter) []readline.PrefixCompleterInterface {
	pciList := []readline.PrefixCompleterInterface{}

	RootCommand.Flags().VisitAll(func(f *pflag.Flag) {
		pciList = append(pciList, readline.PcItem("-"+f.Shorthand, comp.PeerList(env, false)...))
		pciList = append(pciList, readline.PcItem("--"+f.Name, comp.PeerList(env, false)...))
	})

	pciList = append(pciList, comp.PeerList(env, false)...)

	return []readline.PrefixCompleterInterface{readline.PcItem(RootCommand.Use, pciList...)}
}

func receive(cmd *cobra.Command, args []string) {
	defer func() {
		flagWait = false
	}()

	if len(args) != 1 {
		fmt.Println(cmd.Short)
		return
	}

	env := cmd.Context().Value(diameter.EnvContext).(*diameter.Diameter)

	args[0] = peer.NameToId(args[0])
	if _, err := cli.Receive2(env, args[0], flagWait); !diwe.IsIgnorable(err) {
		fmt.Println(err)
	}
}

// Init
//

func init() {
	RootCommand.Flags().BoolVarP(&flagWait, "wait", "w", false, "wait for data")
}
