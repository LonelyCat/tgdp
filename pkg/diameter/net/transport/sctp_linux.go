//
// Project: TGDP - Traffic Generator for Diameter Protocol
// Description: Simple tool for testing and debugging the Diameter protocol
//
// Author: Alexander Kefeli <alexander.kefeli@gmail.com>
//
// File: sctp_linux.go
// Description: Diameter pkg: SCTP transport implementation for Linux
//

package transport

import (
	"fmt"
	"net"
	"net/netip"

	sctp "github.com/georgeyanev/go-sctp"
)

// Methods
//
// SCTP client transport
func (t *Sctp) Connect(remoteAddr netip.Addr, remotePort int, localAddr netip.Addr, localPort int) error {
	rAddr := &sctp.SCTPAddr{
		IPAddrs: []net.IPAddr{
			{IP: remoteAddr.AsSlice()},
		},
		Port: remotePort,
	}

	lAddr := &sctp.SCTPAddr{
		IPAddrs: []net.IPAddr{
			{IP: localAddr.AsSlice()},
		},
		Port: localPort,
	}

	for {
		if conn, err := sctp.DialSCTP("sctp", lAddr, rAddr); err != nil {
			return err
		} else {
			t.Connection = conn
			break
		}
	}

	return nil
}

// SCTP server listener
//
// Create starts an SCTP listener on the given address.
// The address format is "host:port".
func (l *SctpListener) Create(listenAddr string) error {
	addr, err := sctp.ResolveSCTPAddr("sctp", listenAddr)
	if err != nil {
		return err
	}

	if l.listener, err = sctp.ListenSCTP("sctp", addr); err != nil {
		return err
	}

	l.uri = fmt.Sprintf("sctp://%s", listenAddr)
	return nil
}
