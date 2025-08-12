//
// Project: TGDP - Traffic Generator for Diameter Protocol
// Description: Simple tool for testing and debugging the Diameter protocol
//
// Author: Alexander Kefeli <alexander.kefeli@gmail.com>
//
// File: main.go
// Description: Main entry point
//

package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"

	"tgdp/internal/cli"
	"tgdp/internal/flags"
	"tgdp/internal/lua"
	"tgdp/internal/repl"
	"tgdp/pkg/diameter"
	"tgdp/pkg/diameter/net/node"
	"tgdp/pkg/diameter/net/server"
	"tgdp/pkg/diameter/net/sigh"
)

// -- Variables
// --
var (
	version   = "0.0.0"
	buildDate = "2025-01-01 00:00:00"
)

// -- Functions
// --
func usage() {
	fmt.Printf("Usage: %s [flags] [<peer> <app> <command> [<command> ...]]\n", os.Args[0])
	fmt.Printf("       %s [-c <string>] @<Lua script> [args]\n", os.Args[0])
	fmt.Printf("       %s [-c <string>] -y\n", os.Args[0])
	fmt.Println("  <peer>    - Name of peer (must be present in 'node.yaml')")
	fmt.Println("  <app>     - Diameter application NAME or ID")
	fmt.Println("  <command> - application command short NAME or CODE")
	fmt.Println("Flags:")
	flag.PrintDefaults()
	os.Exit(255)
}

func main() {
	fmt.Println("Traffic Generator for Diameter Protocol")
	fmt.Printf("Version: %s\n", version)
	fmt.Printf("Build date: %s\n", buildDate)
	fmt.Println()

	flag.Parse()
	if *flags.H {
		usage()
	}

	diameter.SetVerboseLevel(*flags.V)
	if err := diameter.LoadDictionary(); err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

	if err := node.LoadPeers(); err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

	if *flags.D {
		diameter.Dict.Show()
		return
	}

	if *flags.Y {
		if diameter.Dict.Verify() > 0 {
			os.Exit(1)
		} else {
			return
		}
	}

	defer func() {
		node.DisconnectAll()
		server.Stop()
	}()

	if *flags.S != "" {
		sigh.SetServerMode()
		server.Start(*flags.S, true, nil)
		return
	}

	if flag.NArg() > 0 && flag.Args()[0][0] == '@' {
		lua.Run(flag.Args())
		return
	}

	if flag.NArg() == 0 {
		repl.Run()
		return
	}

	if flag.NArg() < 3 {
		usage()
	}

	cli.Run(flag.Args())
}
