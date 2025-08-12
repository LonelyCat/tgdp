//
// Project: TGDP - Traffic Generator for Diameter Protocol
// Description: Simple tool for testing and debugging the Diameter protocol
//
// Author: Alexander Kefeli <alexander.kefeli@gmail.com>
//
// File: send.go
// Description: REPL: 'send' command implementation
//

package send

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
		Use:   "send",
		Short: "send [-w] <request | answer> <peer> <app> <msg> [<msg> ...]",
		Long:  "Send a message[s] to a peer",
	}

	SubCommandReq = &cobra.Command{
		Use:     "request",
		Aliases: []string{"req"},
		Short:   "send req[uest] <peer> <app> <msg> [<msg> ...]",
		Long:    "Send a request message[s] to a peer",
		Example: "send -w request HSS S6a UL",
		Run:     request,
	}

	SubCommandAns = &cobra.Command{
		Use:     "answer",
		Aliases: []string{"ans"},
		Short:   "send ans[wer] <peer> <app> <msg> [<msg> ...]",
		Long:    "Send an answer message[s] to a peer",
		Example: "send answer DRA 0 dw",
		Run:     answer,
	}
)

// -- Functions
// --
func CompList() []readline.PrefixCompleterInterface {
	pciList := []readline.PrefixCompleterInterface{}

	pciSub := []readline.PrefixCompleterInterface{}
	for _, sub := range RootCommand.Commands() {
		pciSub = append(pciSub, readline.PcItem(sub.Use, comp.PeerList(true)...))
	}

	RootCommand.PersistentFlags().VisitAll(func(f *pflag.Flag) {
		pciList = append(pciList, readline.PcItem("-"+f.Shorthand, pciSub...))
		pciList = append(pciList, readline.PcItem("--"+f.Name, pciSub...))
	})

	pciList = append(pciList, pciSub...)
	pciList = []readline.PrefixCompleterInterface{readline.PcItem(RootCommand.Use, pciList...)}

	return pciList
}

func init() {
	RootCommand.AddCommand(SubCommandReq)
	RootCommand.AddCommand(SubCommandAns)

	RootCommand.PersistentFlags().BoolP(flagWait, "w", false, "wait for data")
}

func request(cmd *cobra.Command, args []string) {
	send(cmd, args, true)
}

func answer(cmd *cobra.Command, args []string) {
	send(cmd, args, false)
}

func send(cmd *cobra.Command, args []string, request bool) {
	if len(args) < 3 {
		fmt.Println(cmd.Short)
		return
	}

	args[0] = list.PeerNameById(args[0])
	if err := cli.Send(args, request, false); err != nil {
		return
	}

	fw, _ := cmd.Flags().GetBool(flagWait)
	for fw {
		if _, err := cli.Receive(args[0], true); err != nil {
			break
		}
	}

	f := cmd.Flags().Lookup(flagWait)
	f.Value.Set(f.DefValue) //nolint:errcheck
}
