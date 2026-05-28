//
// Project: TGDP - Traffic Generator for Diameter Protocol
// Description: Simple tool for testing and debugging the Diameter protocol
//
// Author: Alexander Kefeli <alexander.kefeli@gmail.com>
//
// File: sctp_darwin.go
// Description: Diameter pkg: MacOS not supports SCTP
//

package transport

import (
	"net/netip"
	"tgdp/pkg/diameter/diwe"
)

// Methods
//
// SCTP client transport
func (t *Sctp) Connect(remoteAddr netip.Addr, remotePort int, localAddr netip.Addr, localPort int) error {
	t.Connection = nil
	return &diwe.ErrNotImplemented{}
}

// SCTP server listener
func (l *SctpListener) Create(listenAddr string) error {
	l.listener = nil
	return &diwe.ErrNotImplemented{}
}
