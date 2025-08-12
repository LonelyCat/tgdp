//
// Project: TGDP - Traffic Generator for Diameter Protocol
// Description: Simple tool for testing and debugging the Diameter protocol
//
// Author: Alexander Kefeli <alexander.kefeli@gmail.com>
//
// File: sigh.go
// Description: Diameter pkg: signal handler for graceful shutdown
//

//go:build linux || darwin

package sigh

import (
	"os"
	"os/signal"
	"syscall"
	"tgdp/pkg/diameter/net/node"
	"tgdp/pkg/diameter/net/server"
)

// -- Variables
// --
var (
	serverMode bool
	sigChan    chan os.Signal
)

// -- Functions
// --
func init() {
	sigChan = make(chan os.Signal, 2)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go sigHandler(sigChan)
}

func sigHandler(sigChan chan os.Signal) {
	for {
		s, ok := <-sigChan
		if !ok {
			return
		}
		if serverMode {
			server.Stop()
			continue
		}

		switch s {
		case syscall.SIGINT, syscall.SIGTERM:
			for node := range node.Iter() {
				if node.IsReceiving() {
					node.SendIntSignal()
				}
			}
		}
	}
}

func SetServerMode() {
	serverMode = true
}

func SigChan() chan os.Signal {
	return sigChan
}
