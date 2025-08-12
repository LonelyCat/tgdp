//
// Project: TGDP - Traffic Generator for Diameter Protocol
// Description: Simple tool for testing and debugging the Diameter protocol
//
// Author: Alexander Kefeli <alexander.kefeli@gmail.com>
//
// File: tcp.go
// Description: Diameter pkg: net transport abstractions
//


package transport

import (
	"net"
	"strings"
)

// -- Consts
// --
const DEFAULT_TIMEOUT = 30
const DEFAULT_PORT = 3868
const DEFAULT_PROTOCOL = "sctp"

// -- Types
// --
type ITransport interface {
	// *Sctp | *Tcp

	Connect(net.IP, int, net.IP, int) error
	Close() error
	SetTimeout(int) error

	Send(buf []byte) error
	Recv() ([]byte, error)

	IsConnected() bool
	RemoteAddr() string
	LocalAddr() string
	RemoteIp() net.IP
	LocalIp() net.IP
	RemotePort() int
	LocalPort() int

	Error() error

	Name() string
}

// -- Functions
// --
func New(proto string) (ITransport, error) {
	switch strings.ToLower(proto) {
	case "sctp":
		return &Sctp{}, nil
	case "tcp":
		return &Tcp{}, nil
	default:
		return nil, &ErrUnknownProto{Proto: proto}
	}
}
