//
// Project: TGDP - Traffic Generator for Diameter Protocol
// Description: Simple tool for testing and debugging the Diameter protocol
//
// Author: Alexander Kefeli <alexander.kefeli@gmail.com>
//
// File: cli.go
// Description: Command Line Interface implementation
//

package cli

import (
	"fmt"
	"log/slog"

	"tgdp/internal/flags"
	"tgdp/pkg/diameter"
	"tgdp/pkg/diameter/net/node"
	"tgdp/pkg/diameter/pcap"
)

// -- Variables
// --
var ()

// -- Functions
// --
func Run(args []string) {
	_ = Send(args, true, true)
}

func Send(args []string, request, recv bool) error {
	peer, err := node.GetByName(args[0])
	if err != nil {
		slog.Error(err.Error())
		return err
	}

	app, err := diameter.Dict.GetApp(args[1])
	if err != nil {
		slog.Error(err.Error())
		return err
	}

	cmds := make([]*diameter.Cmd, 0)
	for i := 2; i < len(args); i++ {
		cmd, err := diameter.Dict.GetCmd(args[i], app)
		if err != nil {
			slog.Error(err.Error())
			return nil
		}
		cmds = append(cmds, cmd)
	}

	if !OfflineMode() && !peer.IsConnected() {
		if err = peer.Connect(true); err != nil {
			slog.Error(err.Error())
			return err
		}
		defer peer.Disconnect(true, true) //nolint:errcheck
		diameter.Verbose(peer, diameter.VerbosePeer)
	}

	for i, cmd := range cmds {
		msg, err := diameter.NewMessage(app, cmd, request, true)
		if err != nil {
			slog.Error(err.Error())
			break
		}
		diameter.Verbose(msg, diameter.VerboseMsg)

		if _, err := msg.Serialize(); err != nil {
			slog.Error(err.Error())
			break
		}

		if err := pcap.Write(*flags.W, *flags.A || i > 0, msg.Buff(), peer, msg.IsRequest()); err != nil {
			slog.Error(err.Error())
		}

		if !OfflineMode() {
			if err = peer.SendTo2(msg.Buff()); err != nil {
				slog.Error(err.Error())
				break
			}

			if recv {
				if _, err := Receive(peer.Name, true); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func Receive(name string, wait bool) (*diameter.Message, error) {
	peer, err := node.GetByName(name)
	if err != nil {
		slog.Error(err.Error())
		return nil, err
	}

	if peer.HasData() || wait {
		msg, err := peer.RecvFrom()
		if err != nil {
			switch err.(type) {
			case *node.ErrInterrupted:
				fmt.Printf("\n%v\n", err)
			default:
				slog.Error(err.Error())
			}
			return nil, err
		}
		diameter.Verbose(msg, diameter.VerboseMsg)

		if err := pcap.Write(*flags.W, true, msg.Buff(), peer, msg.IsRequest()); err != nil {
			slog.Error(err.Error())
		}

		return msg, nil
	}

	return nil, &node.ErrNoData{Peer: peer.Name}
}

func OfflineMode() bool {
	return *flags.N
}
