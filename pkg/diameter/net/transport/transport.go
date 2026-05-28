//
// Project: TGDP - Traffic Generator for Diameter Protocol
// Description: Simple tool for testing and debugging the Diameter protocol
//
// Author: Alexander Kefeli <alexander.kefeli@gmail.com>
//
// File: transport.go
// Description: Diameter pkg: net transport abstractions
//

package transport

import (
	"errors"
	"io"
	"io/fs"
	"net"
	"net/netip"
	"os"
	"strings"
	"syscall"
)

// Consts
//

const DefaultTimeout = 10
const DefaultPort = 3868
const DefaultProtocol = "sctp"

// Types
//

// ITransport is the interface for network transports (SCTP or TCP).
// It abstracts the transport layer, allowing the client to work with either protocol.
type ITransport interface {
	// *Sctp | *Tcp

	Connect(netip.Addr, int, netip.Addr, int) error
	Close() error
	SetTimeout(int) error

	Send(buf []byte) error
	Recv() ([]byte, error)

	IsConnected() bool
	RemoteAddr() string
	LocalAddr() string
	RemoteIp() netip.Addr
	LocalIp() netip.Addr
	RemotePort() int
	LocalPort() int

	Error() error

	Name() string
	Type() int
}

// IListener is an interface for network listeners (SCTP or TCP).
// It abstracts the transport layer, allowing the server to work with either protocol.
type IListener interface {
	Create(string) error
	Accept() (ITransport, error)
	Close() error
	Ready() bool
	Name() string
	Uri() string
}

// Constructors
//

// New creates a new ITransport instance based on the given protocol.
// Supported protocols are "sctp" and "tcp".
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

// Helpers
//
// isClosedError returns true if the error is a closed error (net.ErrClosed, io.EOF, syscall.EPIPE, ...).
func IsClosedError(err error) bool {
	if _, ok := err.(*net.OpError); ok {
		return true
	}
	return errors.Is(err, net.ErrClosed) ||
		errors.Is(err, fs.ErrClosed) ||
		errors.Is(err, os.ErrClosed) ||
		errors.Is(err, syscall.ENOTCONN) ||
		errors.Is(err, syscall.ECONNRESET) ||
		errors.Is(err, syscall.EINVAL) ||
		errors.Is(err, syscall.EPIPE) ||
		errors.Is(err, io.EOF)
}
