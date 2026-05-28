//
// Project: TGDP - Traffic Generator for Diameter Protocol
// Description: Simple tool for testing and debugging the Diameter protocol
//
// Author: Alexander Kefeli <alexander.kefeli@gmail.com>
//
// File: server.go
// Description: CLI server handling
//

package server

import (
	"log/slog"
	"os"
	"os/signal"

	"tgdp/internal/flags"

	"tgdp/pkg/diameter"
	ds "tgdp/pkg/diameter/net/server"
)

func Run(d *diameter.Diameter) {
	server := ds.New(d)
	server.SetAutoreply(true)
	server.SetVerboseLevel(ds.Info)

	if err := server.Start(*flags.S, true); err != nil {
		slog.Error(err.Error())
		return
	}
	server.Dump()

	ccChan := make(chan os.Signal, 1)
	signal.Notify(ccChan, os.Interrupt)
	<-ccChan
	server.Shutdown()
}
