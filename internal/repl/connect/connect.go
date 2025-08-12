//
// Project: TGDP - Traffic Generator for Diameter Protocol
// Description: Simple tool for testing and debugging the Diameter protocol
//
// Author: Alexander Kefeli <alexander.kefeli@gmail.com>
//
// File: connect.go
// Description: REPL: 'connect' command implementation
//

package connect

import (
	"fmt"
	"strconv"
	"strings"

	"tgdp/internal/repl/comp"
	"tgdp/internal/repl/list"
	"tgdp/pkg/diameter/net/node"
	"tgdp/pkg/diameter/net/transport"

	"github.com/chzyer/readline"
	"github.com/spf13/cobra"
)

// -- Variables
// --
var (
	RootCommand = &cobra.Command{
		Use:     "connect",
		Short:   "connect <peer | address [port] [transport]>",
		Long:    "Connect to a peer",
		Example: "connect MME",
		Run:     connect,
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

func connect(cmd *cobra.Command, args []string) {
	if len(args) < 1 {
		fmt.Println(cmd.Short)
		return
	}

	if len(args) == 1 {
		peer, err := node.GetByName(list.PeerNameById(args[0]))
		if err != nil {
			connectAddress(args)
		} else {
			connectPeer(peer)
		}
	} else {
		connectAddress(args)
	}
}

func connectPeer(peer *node.Node) {
	if err := peer.Connect(true); err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("Connected to peer '%s'\n", peer.Name)
}

func connectAddress(args []string) {
	address := args[0]
	port := transport.DEFAULT_PORT
	proto := transport.DEFAULT_PROTOCOL

	if len(args) > 1 {
		p, err := strconv.Atoi(args[1])
		if err != nil {
			fmt.Println(err)
			return
		}
		port = p
	}

	if len(args) > 2 {
		proto = strings.ToLower(args[2])
	}

	name := fmt.Sprintf("peer-%s:%d", address, port)
	node, err := node.New(name, address, port, proto, transport.DEFAULT_TIMEOUT)
	if err != nil {
		fmt.Printf("%v\n", err)
		return
	}

	err = node.Connect(true)
	if err != nil {
		fmt.Printf("%v\n", err)
		return
	}

	fmt.Printf("Connected to %s://%s:%d\n", proto, address, port)
}
