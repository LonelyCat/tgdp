//
// Project: TGDP - Traffic Generator for Diameter Protocol
// Description: Simple tool for testing and debugging the Diameter protocol
//
// Author: Alexander Kefeli <alexander.kefeli@gmail.com>
//
// File: avp.go
// Description: REPL: 'pcap' command implementation
//

package pcap

import (
	"fmt"
	"tgdp/pkg/diameter"

	"github.com/chzyer/readline"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// Variables
//

var (
	RootCommand = &cobra.Command{
		Use:   "pcap",
		Short: "pcap <open <file> | close | status>",
		Long:  "Save messages to a PCAP file",
	}

	SubCommandOpen = &cobra.Command{
		Use:     "open",
		Short:   "pcap open [-t] <file>",
		Long:    "Enable saving to the PCAP file",
		Example: "pcap open trace.pcap",
		Run:     open,
	}

	SubCommandClose = &cobra.Command{
		Use:     "close",
		Short:   "pcap close",
		Long:    "Disable saving to PCAP",
		Example: "pcap close",
		Run:     close,
	}

	SubCommandStatus = &cobra.Command{
		Use:     "status",
		Short:   "pcap status",
		Long:    "Show PCAP status",
		Example: "pcap status",
		Run:     status,
	}
)

var (
	flagTruncate bool
)

// Functions
//

func CompList(env *diameter.Diameter) []readline.PrefixCompleterInterface {
	pciSub := []readline.PrefixCompleterInterface{}

	for _, sub := range RootCommand.Commands() {
		subFlags := []readline.PrefixCompleterInterface{}
		sub.Flags().VisitAll(func(f *pflag.Flag) {
			subFlags = append(subFlags, readline.PcItem("-"+f.Shorthand))
			subFlags = append(subFlags, readline.PcItem("--"+f.Name))
		})
		pciSub = append(pciSub, readline.PcItem(sub.Use, subFlags...))
	}

	return []readline.PrefixCompleterInterface{readline.PcItem(RootCommand.Use, pciSub...)}
}

func open(cmd *cobra.Command, args []string) {
	defer func() {
		flagTruncate = false
	}()

	if len(args) == 0 {
		fmt.Println(cmd.Short)
		return
	}

	env := cmd.Context().Value(diameter.EnvContext).(*diameter.Diameter)
	if env.Pcap() == nil {
		env.NewPcapWriter()
	}

	if err := env.Pcap().Open(args[0], !flagTruncate); err != nil {
		fmt.Println(err)
	}

	status(cmd, nil)
}

func close(cmd *cobra.Command, args []string) {
	env := cmd.Context().Value(diameter.EnvContext).(*diameter.Diameter)
	if env.Pcap() == nil {
		return
	}

	if err := env.Pcap().Close(); err != nil {
		fmt.Println(err)
	}
}

func status(cmd *cobra.Command, args []string) {
	env := cmd.Context().Value(diameter.EnvContext).(*diameter.Diameter)
	if env.Pcap() == nil {
		return
	}

	fmt.Println("PCAP status:")

	fmt.Print("  state: ")
	if env.Pcap().IsOpen() {
		fmt.Println("OPEN")
	} else {
		fmt.Println("CLOSED")
	}

	fmt.Print("   mode: ")
	if env.Pcap().IsAppend() {
		fmt.Println("APPEND")
	} else {
		fmt.Println("TRUNCATE")
	}

	fmt.Printf("   file: %s\n", env.Pcap().File())
}

// Init
//

func init() {
	SubCommandOpen.Flags().BoolVarP(&flagTruncate, "truncate", "t", false, "truncate PCAP file if exists")

	RootCommand.AddCommand(SubCommandOpen)
	RootCommand.AddCommand(SubCommandClose)
	RootCommand.AddCommand(SubCommandStatus)
}
