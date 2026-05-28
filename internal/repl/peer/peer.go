//
// Project: TGDP - Traffic Generator for Diameter Protocol
// Description: Simple tool for testing and debugging the Diameter protocol
//
// Author: Alexander Kefeli <alexander.kefeli@gmail.com>
//
// File: avp.go
// Description: REPL: 'peer' command implementation
//

package peer

import (
	"fmt"
	"strconv"
	"strings"

	"tgdp/internal/repl/comp"
	"tgdp/pkg/diameter"
	"tgdp/pkg/diameter/net/node"
	"tgdp/pkg/diameter/net/transport"

	"github.com/chzyer/readline"
	"github.com/spf13/cobra"
)

// Variables
//

var (
	peersNames []string

	RootCommand = &cobra.Command{
		Use:   "peer",
		Short: "peer <list | info | open | close> [parameters]",
		Long:  "Manage remote peers",
	}

	SubCommandList = &cobra.Command{
		Use:     "list",
		Short:   "peer list",
		Long:    "Show list of peers",
		Example: "peer list",
		Run:     list,
	}

	SubCommandInfo = &cobra.Command{
		Use:     "info",
		Short:   "peer info <id | name>",
		Long:    "Show peer information",
		Example: "peer info HSS",
		Run:     info,
	}
	SubCommandOpen = &cobra.Command{
		Use:     "open",
		Short:   "peer open <name | address:[port]>",
		Long:    "Open a peer connection",
		Example: "peer open HSS",
		Run:     open,
	}

	SubCommandClose = &cobra.Command{
		Use:     "close",
		Short:   "peer close <id | name>",
		Long:    "Close a peer connection",
		Example: "close HSS",
		Run:     close,
	}
)

// Functions
//

func CompList(env *diameter.Diameter) []readline.PrefixCompleterInterface {
	pciSub := []readline.PrefixCompleterInterface{}

	for _, sub := range RootCommand.Commands() {
		switch sub {
		case SubCommandList:
			pciSub = append(pciSub, readline.PcItem(sub.Use))
		default:
			pciSub = append(pciSub, readline.PcItem(sub.Use, comp.PeerList(env, false)...))
		}
	}

	return []readline.PrefixCompleterInterface{readline.PcItem(RootCommand.Use, pciSub...)}
}

func list(cmd *cobra.Command, args []string) {
	env := cmd.Context().Value(diameter.EnvContext).(*diameter.Diameter)

	peersNames = nil
	for i, n := range env.Peers().Iter2() {
		peersNames = append(peersNames, n.Name)
		fmt.Printf("%3d", i)
		if n.IsOpen() {
			fmt.Print(" * ")
		} else {
			fmt.Print("   ")
		}
		fmt.Printf("%s \t%s \t%d \t%s\n", n.Name, n.Address, n.RemotePort, n.Transport().Name())
	}
}

func info(cmd *cobra.Command, args []string) {
	if len(args) == 0 {
		fmt.Println(cmd.Short)
		return
	}

	env := cmd.Context().Value(diameter.EnvContext).(*diameter.Diameter)

	if peer, err := env.Peers().GetByName(args[0]); err != nil {
		fmt.Println(err)
	} else {
		peer.Trace()
	}
}

func open(cmd *cobra.Command, args []string) {
	if len(args) < 1 {
		fmt.Println(cmd.Short)
		return
	}

	env := cmd.Context().Value(diameter.EnvContext).(*diameter.Diameter)

	if len(args) == 1 {
		peer, err := env.Peers().GetByName(NameToId(args[0]))
		if err != nil {
			connectAddress(env, args)
		} else {
			connectPeer(peer)
		}
	} else {
		connectAddress(env, args)
	}
}

func close(cmd *cobra.Command, args []string) {
	if len(args) != 1 {
		fmt.Println(cmd.Short)
		return
	}

	env := cmd.Context().Value(diameter.EnvContext).(*diameter.Diameter)

	peer, err := env.Peers().GetByName(NameToId(args[0]))
	if err != nil {
		fmt.Println(err)
		return
	}

	if err := peer.Disconnect(); err != nil {
		fmt.Println(err)
		return
	}

	if peer.IsClient() {
		env.Peers().Remove(peer.Name)
	}

	fmt.Printf("Closed connection to '%s'\n", peer.Name)
}

// Helpers
//

func NameToId(id string) string {
	i, err := strconv.Atoi(id)
	if err == nil && i < len(peersNames) {
		return peersNames[i]
	}
	return id
}

func connectPeer(peer *node.Node) {
	if err := peer.Connect(); err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("Connected to peer '%s'\n", peer.Name)
}

func connectAddress(env *diameter.Diameter, args []string) {
	address := args[0]
	port := transport.DefaultPort
	proto := transport.DefaultProtocol

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
	node, err := env.Peers().NewPeer(name, address, port, proto, transport.DefaultTimeout, env)
	if err != nil {
		fmt.Printf("%v\n", err)
		return
	}

	err = node.Connect()
	if err != nil {
		fmt.Printf("%v\n", err)
		return
	}

	fmt.Printf("Connected to %s://%s:%d\n", proto, address, port)
}

// Init
//

func init() {
	RootCommand.AddCommand(SubCommandList)
	RootCommand.AddCommand(SubCommandInfo)
	RootCommand.AddCommand(SubCommandOpen)
	RootCommand.AddCommand(SubCommandClose)
}
