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
	"tgdp/internal/repl/peer"
	"tgdp/pkg/diameter"

	"github.com/chzyer/readline"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// Variables
//

var (
	RootCommand = &cobra.Command{
		Use:   "send",
		Short: "send  <request [-w | --wait] | answer> <peer> <app> <msg> [<msg> ...]",
		Long:  "Send a message[s] to a peer",
	}

	SubCommandReq = &cobra.Command{
		Use:     "request",
		Aliases: []string{"req"},
		Short:   "send req[uest] [-w | --wait] <peer> <app> <msg> [<msg> ...]",
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

var (
	flagWait bool
)

// Functions
//

//   59 │ func CompList() []readline.PrefixCompleterInterface {
//   60 │     pciList := []readline.PrefixCompleterInterface{}
//   61 │
//   62 │     pciSub := []readline.PrefixCompleterInterface{}
//   63 │     for _, sub := range RootCommand.Commands() {
//   64 │         pciSub = append(pciSub, readline.PcItem(sub.Use, comp.PeerList(true)...))
//   65 │     }
//   66 │
//   67 │     RootCommand.PersistentFlags().VisitAll(func(f *pflag.Flag) {
//   68 │         pciList = append(pciList, readline.PcItem("-"+f.Shorthand, pciSub...))
//   69 │         pciList = append(pciList, readline.PcItem("--"+f.Name, pciSub...))
//   70 │     })
//   71 │
//   72 │     pciList = append(pciList, pciSub...)
//   73 │     pciList = []readline.PrefixCompleterInterface{readline.PcItem(RootCommand.Use, pciList...)}
//   74 │
//   75 │     return pciList
//   76 │ }

func CompList(env *diameter.Diameter) []readline.PrefixCompleterInterface {
	pciPeers := comp.PeerList(env, true)

	pciSub := []readline.PrefixCompleterInterface{}
	for _, sub := range RootCommand.Commands() {
		subFlags := make([]readline.PrefixCompleterInterface, 0)
		sub.Flags().VisitAll(func(f *pflag.Flag) {
			subFlags = append(subFlags, readline.PcItem("-"+f.Shorthand, pciPeers...))
			subFlags = append(subFlags, readline.PcItem("--"+f.Name, pciPeers...))
		})
		if len(subFlags) > 0 {
			pciSub = append(pciSub, readline.PcItem(sub.Use, append(subFlags, pciPeers...)...))
		} else {
			pciSub = append(pciSub, readline.PcItem(sub.Use, pciPeers...))
		}
	}

	return []readline.PrefixCompleterInterface{readline.PcItem(RootCommand.Use, pciSub...)}
}

func request(cmd *cobra.Command, args []string) {
	send(cmd, args, true)
}

func answer(cmd *cobra.Command, args []string) {
	send(cmd, args, false)
}

func send(cmd *cobra.Command, args []string, request bool) {
	defer func() {
		flagWait = false
	}()

	if len(args) < 3 {
		fmt.Println(cmd.Short)
		return
	}

	env := cmd.Context().Value(diameter.EnvContext).(*diameter.Diameter)

	args[0] = peer.NameToId(args[0])
	if err := cli.Send(env, args, request, flagWait, false); err != nil {
		return
	}
}

// Init
//

func init() {
	SubCommandReq.Flags().BoolVarP(&flagWait, "wait", "w", false, "wait for data")

	RootCommand.AddCommand(SubCommandReq)
	RootCommand.AddCommand(SubCommandAns)
}
