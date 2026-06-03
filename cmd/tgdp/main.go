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
	"tgdp/internal/config"
	"tgdp/internal/flags"
	"tgdp/internal/lua"
	"tgdp/internal/repl"
	"tgdp/internal/server"
	"tgdp/internal/version"

	"tgdp/pkg/diameter"
	"tgdp/pkg/diameter/dict"
)

// Functions
//

func exitOnError(err error) {
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
}

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
	flag.Parse()
	if *flags.H {
		usage()
	}

	if *flags.Version {
		version.Show()
	}

	if err := config.Load(*flags.C); err != nil {
		fmt.Println(err)
		fmt.Println("Default settings applied")
	}

	d, err := diameter.New(config.DiaMode())
	exitOnError(err)

	if *flags.W != "" {
		d.NewPcapWriter()
		exitOnError(d.PcapOpen(*flags.W, *flags.A))
	}

	exitOnError(d.LoadDict(config.DialDictFile(), dict.FormatPkl))

	if *flags.D {
		d.Dict().Show()
		return
	}

	if *flags.Y {
		if d.Dict().Verify() > 0 {
			os.Exit(1)
		} else {
			return
		}
	}

	exitOnError(d.LoadData(config.AvpsDataFile()))
	exitOnError(d.LoadPeers(config.PeersDataFile()))

	d.SetTraceLevel(int32(*flags.V))

	defer func() {
		d.PcapClose() // nolint: errcheck
	}()

	if *flags.S != "" {
		server.Run(d)
		return
	}

	if flag.NArg() > 0 && flag.Args()[0][0] == '@' {
		lua.Run(d, flag.Args())
		return
	}

	if flag.NArg() == 0 {
		repl.Run(d)
		return
	}

	if flag.NArg() < 3 {
		usage()
	}

	cli.Run(d, flag.Args())
}
