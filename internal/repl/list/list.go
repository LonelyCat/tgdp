//
// Project: TGDP - Traffic Generator for Diameter Protocol
// Description: Simple tool for testing and debugging the Diameter protocol
//
// Author: Alexander Kefeli <alexander.kefeli@gmail.com>
//
// File: list.go
// Description: REPL: 'list' command implementation
//

package list

import (
	"fmt"
	"sort"
	"strconv"

	"tgdp/pkg/diameter"
	"tgdp/pkg/diameter/net/node"

	"github.com/chzyer/readline"
	"github.com/spf13/cobra"
)

// -- Variables
// --
var (
	peersNames []string

	RootCommand = &cobra.Command{
		Use:     "list",
		Aliases: []string{"avps", "peers"},
		Short:   "list <avps | peers>",
		Long:    "Show list of objects",
	}

	SubCommandPeers = &cobra.Command{
		Use:     "peers",
		Short:   "list peers",
		Long:    "List peers or connections",
		Example: "list peers",
		Run:     peers,
	}

	SubCommandAvps = &cobra.Command{
		Use:     "avps",
		Short:   "list avps",
		Long:    "List AVPs and values",
		Example: "list avps",
		Run:     avps,
	}
)

// -- Functions
// --
func CompList() []readline.PrefixCompleterInterface {
	pciSub := []readline.PrefixCompleterInterface{}
	for _, sub := range RootCommand.Commands() {
		pciSub = append(pciSub, readline.PcItem(sub.Use))
	}

	return []readline.PrefixCompleterInterface{readline.PcItem(RootCommand.Use, pciSub...)}
}

func init() {
	RootCommand.AddCommand(SubCommandAvps)
	RootCommand.AddCommand(SubCommandPeers)
}

func peers(cmd *cobra.Command, args []string) {
	peersNames = nil
	for i, n := range node.Iter2() {
		peersNames = append(peersNames, n.Name)
		fmt.Printf("%3d", i)
		if n.IsConnected() {
			fmt.Print(" * ")
		} else {
			fmt.Print("   ")
		}
		fmt.Printf("%s \t%s \t%d \t%s\n", n.Name, n.Address, n.RemotePort, n.Tr.Name())
	}
}

func PeerNameById(id string) string {
	i, err := strconv.Atoi(id)
	if err == nil && i < len(peersNames) {
		return peersNames[i]
	}
	return id
}

func avps(cmd *cobra.Command, args []string) {
	listAvpsData(nil, 0)
}

func listAvpsData(store *diameter.AvpDataStore, shift int) {
	type avpData struct {
		avp    *diameter.Avp
		values []*diameter.AvpData
	}

	list := make([]avpData, 0)

	for code, data := range diameter.AvpDataIter2(store) {
		avp, err := diameter.Dict.GetAvpByCode(code)
		if err != nil {
			continue
		}

		list = append(list, avpData{
			avp:    avp,
			values: data,
		})
	}

	sort.Slice(list, func(i, j int) bool {
		return list[i].avp.Name < list[j].avp.Name
	})

	for _, data := range list {
		for _, value := range data.values {
			for range shift {
				fmt.Print(" ")
			}
			fmt.Printf("%s (%d)", data.avp.Name, data.avp.Code)
			if data.avp.IsGrouped() {
				fmt.Println()
				listAvpsData(value.Value.(*diameter.AvpDataStore), shift+2)
			} else {
				fmt.Printf(" = %v\n", value.Value)
			}
		}
	}
}
