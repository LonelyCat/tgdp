//
// Project: TGDP - Traffic Generator for Diameter Protocol
// Description: Simple tool for testing and debugging the Diameter protocol
//
// Author: Alexander Kefeli <alexander.kefeli@gmail.com>
//
// File: close.go
// Description: REPL: 'close' command implementation
//

package close

import (
	"fmt"

	"tgdp/internal/repl/comp"
	"tgdp/internal/repl/list"
	"tgdp/pkg/diameter/net/node"

	"github.com/chzyer/readline"
	"github.com/spf13/cobra"
)

// -- Variables
// --
var (
	RootCommand = &cobra.Command{
		Use:     "close",
		Short:   "close <peer>",
		Long:    "Close connection to a peer",
		Example: "close DRA",
		Run:     close,
	}
)

// -- Functions
// --
func CompList() []readline.PrefixCompleterInterface {
	pci := readline.PcItem(
		RootCommand.Use,
		comp.PeerList(false)...,
	)

	return []readline.PrefixCompleterInterface{pci}
}

func close(cmd *cobra.Command, args []string) {
	if len(args) != 1 {
		fmt.Println(cmd.Short)
		return
	}

	peer, err := node.GetByName(list.PeerNameById(args[0]))
	if err != nil {
		fmt.Println(err)
		return
	}

	if err := peer.Disconnect(true, false); err != nil {
		fmt.Println(err)
		return
	}

	if peer.IsClient() {
		node.Remove(peer.Name)
	}

	fmt.Printf("Closed connection to '%s'\n", peer.Name)
}
