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
	"tgdp/pkg/diameter/diwe"
	"tgdp/pkg/diameter/net/node"
	"tgdp/pkg/diameter/pcap"
)

// Functions
//

func Run(env *diameter.Diameter, args []string) {
	_ = Send(env, args, true, true, true)
}

func Send(env *diameter.Diameter, args []string, request, recv bool, disconnect bool) error {
	peer, err := env.Peers().GetByName(args[0])
	if err != nil {
		slog.Error(err.Error())
		return err
	}

	if !OfflineMode() && !peer.IsOpen() {
		if err = peer.Connect(); err != nil {
			slog.Error(err.Error())
			return err
		}
		env.Trace(peer, diameter.TracePeer)
	}

	for _, cmd := range args[2:] {
		msg, err := env.NewMessage(args[1], cmd, request, true)
		if err != nil {
			slog.Error(err.Error())
			break
		}
		env.Trace(msg, diameter.TraceMsg)

		_, err = msg.Serialize()
		if err != nil {
			slog.Error(err.Error())
			break
		}

		err = env.Pcap().Write(msg.Bytes(), peer, pcap.DirOutgoing)
		if err != nil {
			slog.Error(err.Error())
		}
		env.Pcap().Append(pcap.Append)

		if !OfflineMode() {
			if err = env.SendMessage(peer, msg); err != nil {
				slog.Error(err.Error())
				break
			}

			if recv {
				if _, err := Receive(env, peer, true); err != nil {
					return err
				}
			}
		}
	}

	if !OfflineMode() && disconnect {
		if err := peer.Disconnect(); err != nil {
			slog.Error(err.Error())
		}
	}

	return nil
}

func Receive(env *diameter.Diameter, peer *node.Node, wait bool) (*diameter.Message, error) {
	if peer.HasData() || wait {
		msg, err := env.RecvMessage(peer, wait)
		if err != nil {
			switch err.(type) {
			case *diwe.InfInterrupted:
				fmt.Printf("\n%v\n", err)
			default:
				slog.Error(err.Error())
			}
			return nil, err
		}
		env.Trace(msg, diameter.TraceMsg)

		if err := env.Pcap().Write(msg.Bytes(), peer, pcap.DirIncoming); err != nil {
			slog.Error(err.Error())
		}

		return msg, nil
	}

	return nil, &diwe.ErrNoData{Peer: peer.Name}
}

func Receive2(env *diameter.Diameter, peerName string, wait bool) (*diameter.Message, error) {
	peer, err := env.Peers().GetByName(peerName)
	if err != nil {
		return nil, err
	}

	return Receive(env, peer, wait)
}

func OfflineMode() bool {
	return *flags.N
}
