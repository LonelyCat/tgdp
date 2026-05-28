//
// Project: TGDP - Traffic Generator for Diameter Protocol
// Description: Simple tool for testing and debugging the Diameter protocol
//
// Author: Alexander Kefeli <alexander.kefeli@gmail.com>
//
// File: sctp.go
// Description: Diameter pkg: SCTP transport implementation
//

package transport

import (
	"net/netip"
	"time"

	sctp "github.com/georgeyanev/go-sctp"
)

// Consts
//

const (
	TransportSctp = 1
)

// Types
//
// SCTP client transport
type Sctp struct {
	Connection *sctp.SCTPConn
	Err        error
}

// SctpListener wraps a TCP network listener.
type SctpListener struct {
	uri      string
	listener *sctp.SCTPListener
}

// Methods
//
// SCTP client transport
func (t *Sctp) Close() error {
	if t != nil && t.Connection != nil {
		err := t.Connection.Close()
		// t.Connection = nil
		t.Err = err
		return err
	}

	return nil
}

func (t *Sctp) SetTimeout(timeout int) error {
	return t.Connection.SetDeadline(time.Now().Add(time.Duration(timeout) * time.Second))
}

func (t *Sctp) Send(buf []byte) error {
	if _, err := t.Connection.Write(buf); err != nil {
		t.Err = err
		return err
	}

	return nil
}

func (t *Sctp) Recv() ([]byte, error) {
	var buf [65535]byte

	if n, err := t.Connection.Read(buf[:]); err == nil {
		data := make([]byte, n)
		copy(data, buf[:n])
		return data, nil
	} else {
		t.Err = err
		return nil, err
	}
}

func (t *Sctp) IsConnected() bool {
	return t.Connection != nil
}

func (t *Sctp) RemoteAddr() string {
	if t.IsConnected() {
		return t.Connection.RemoteAddr().String()
	}

	return ""
}

func (t *Sctp) LocalAddr() string {
	if t.IsConnected() {
		return t.Connection.LocalAddr().String()
	}

	return ""
}

func (t *Sctp) RemoteIp() netip.Addr {
	if t.IsConnected() {
		if addr, ok := netip.AddrFromSlice(t.Connection.RemoteAddr().(*sctp.SCTPAddr).IPAddrs[0].IP); ok {
			return addr
		}
	}

	return netip.Addr{}
}

func (t *Sctp) LocalIp() netip.Addr {
	if t.IsConnected() {
		if addr, ok := netip.AddrFromSlice(t.Connection.LocalAddr().(*sctp.SCTPAddr).IPAddrs[0].IP); ok {
			return addr
		}
	}

	return netip.Addr{}
}

func (t *Sctp) RemotePort() int {
	if t.IsConnected() {
		return t.Connection.RemoteAddr().(*sctp.SCTPAddr).Port
	}
	return 0
}

func (t *Sctp) LocalPort() int {
	if t.IsConnected() {
		return t.Connection.LocalAddr().(*sctp.SCTPAddr).Port
	}

	return 0
}

func (t *Sctp) Error() error {
	return t.Err
}

func (t *Sctp) Name() string {
	return "SCTP"
}

func (t *Sctp) Type() int {
	return TransportSctp
}

// SCTP server listener
//
// Accept waits for and returns the next SCTP connection.
// Returns a transport.ITransport or an error if the listener is closed.
func (l *SctpListener) Accept() (ITransport, error) {
	conn, err := l.listener.AcceptSCTP()
	if err != nil {
		return nil, err
	}

	return &Sctp{Connection: conn, Err: nil}, nil
}

// Close closes the SCTP listener, stopping it from accepting new connections.
func (l *SctpListener) Close() error {
	if l.listener == nil {
		return nil
	}

	return l.listener.Close()
}

// Ready returns true if the SCTP listener has been created and is ready to accept connections.
func (l *SctpListener) Ready() bool {
	return l.listener != nil
}

// Name returns the transport type name "SCTP".
func (l *SctpListener) Name() string {
	return "SCTP"
}

// Uri returns the URI of the SCTP listener in the format "sctp://host:port".
func (l *SctpListener) Uri() string {
	return l.uri
}
