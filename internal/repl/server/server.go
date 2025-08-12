//
// Project: TGDP - Traffic Generator for Diameter Protocol
// Description: Simple tool for testing and debugging the Diameter protocol
//
// Author: Alexander Kefeli <alexander.kefeli@gmail.com>
//
// File: server.go
// Description: REPL: 'server' command implementation
//

package server

import (
	"fmt"

	srv "tgdp/pkg/diameter/net/server"
	"tgdp/pkg/diameter/net/transport"

	"github.com/chzyer/readline"
	"github.com/spf13/cobra"
)

// -- Variables
// --
var (
	RootCommand = &cobra.Command{
		Use:       "server",
		Short:     "server <start |stop | status> [<address> [port]]",
		Long:      "Control Diameter server",
		ValidArgs: []string{"start", "stop", "status"},
	}

	SubCommandStart = &cobra.Command{
		Use:     "start",
		Short:   "server start <address> [port]",
		Long:    "Start Diameter server",
		Example: "server start 127.0.0.1 3868",
		Run:     start,
	}

	SubCommandStop = &cobra.Command{
		Use:     "stop",
		Short:   "server stop",
		Long:    "Stop Diameter server",
		Example: "server stop",
		Run:     stop,
	}

	SubCommandStatus = &cobra.Command{
		Use:     "status",
		Short:   "server status",
		Long:    "Check Diameter server status",
		Example: "server status",
		Run:     status,
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
	RootCommand.AddCommand(SubCommandStart)
	RootCommand.AddCommand(SubCommandStop)
	RootCommand.AddCommand(SubCommandStatus)
}

func start(cmd *cobra.Command, args []string) {
	if len(args) < 1 {
		fmt.Println(cmd.Short)
		return
	}

	listenAddr := func() string {
		if len(args) > 1 {
			return fmt.Sprintf("%s:%s", args[0], args[1])
		} else {
			return fmt.Sprintf("%s:%d", args[0], transport.DEFAULT_PORT)
		}
	}

	sema := make(chan struct{}, 1)
	go srv.Start(listenAddr(), false, &sema)
	<-sema
	srv.ShowState()
}

func stop(cmd *cobra.Command, args []string) {
	srv.Stop()
}

func status(cmd *cobra.Command, args []string) {
	srv.ShowState()
}
