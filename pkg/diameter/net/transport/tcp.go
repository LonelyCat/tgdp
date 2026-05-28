//
// Project: TGDP - Traffic Generator for Diameter Protocol
// Description: Simple tool for testing and debugging the Diameter protocol
//
// Author: Alexander Kefeli <alexander.kefeli@gmail.com>
//
// File: tcp.go
// Description: Diameter pkg: TCP transport implementation
//

package transport

import (
	"fmt"
	"net"
	"net/netip"
	"time"
)

// Consts
//

const (
	TransportTcp = 2
)

// Types
//

// TCP client transport
type Tcp struct {
	Connection *net.TCPConn
	Err        error
}

// TcpListener wraps a TCP network listener.
type TcpListener struct {
	uri      string
	listener *net.TCPListener
}

// Methods
//
// TCP client transport methods
func (t *Tcp) Connect(remoteAddr netip.Addr, remotePort int, localAddr netip.Addr, localPort int) error {
	rAddr := &net.TCPAddr{
		IP:   remoteAddr.AsSlice(),
		Port: remotePort,
	}

	lAddr := &net.TCPAddr{
		IP:   localAddr.AsSlice(),
		Port: localPort,
	}

	conn, err := net.DialTCP("tcp", lAddr, rAddr)
	if err != nil {
		t.Err = err
		return err
	}

	t.Connection = conn

	return nil
}

func (t *Tcp) Close() error {
	if t != nil && t.Connection != nil {
		err := t.Connection.Close()
		t.Err = err
		return err
	}

	return nil
}

func (t *Tcp) SetTimeout(timeout int) error {
	return t.Connection.SetDeadline(time.Now().Add(time.Duration(timeout) * time.Second))
}

func (t *Tcp) Send(buf []byte) error {
	if _, err := (*t.Connection).Write(buf); err != nil {
		t.Err = err
		return err
	}

	return nil
}

func (t *Tcp) Recv() ([]byte, error) {
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

func (t *Tcp) IsConnected() bool {
	return t.Connection != nil
}

func (t *Tcp) RemoteAddr() string {
	if t.IsConnected() {
		return (*t.Connection).RemoteAddr().String()
	}

	return ""
}

func (t *Tcp) LocalAddr() string {
	if t.IsConnected() {
		return (*t.Connection).LocalAddr().String()
	}

	return ""
}

func (t *Tcp) RemoteIp() netip.Addr {
	if t.IsConnected() {
		if addr, ok := netip.AddrFromSlice(t.Connection.RemoteAddr().(*net.TCPAddr).IP); ok {
			return addr
		}
	}

	return netip.Addr{}
}

func (t *Tcp) LocalIp() netip.Addr {
	if t.IsConnected() {
		if addr, ok := netip.AddrFromSlice(t.Connection.LocalAddr().(*net.TCPAddr).IP); ok {
			return addr
		}
	}

	return netip.Addr{}
}

func (t *Tcp) RemotePort() int {
	if t.IsConnected() {
		return int(t.Connection.RemoteAddr().(*net.TCPAddr).Port)
	}

	return 0
}

func (t *Tcp) LocalPort() int {
	if t.IsConnected() {
		return int(t.Connection.LocalAddr().(*net.TCPAddr).Port)
	}

	return 0
}

func (t *Tcp) Error() error {
	return t.Err
}

func (t *Tcp) Name() string {
	return "TCP"
}

func (t *Tcp) Type() int {
	return TransportTcp
}

// TCP server listener methods
//
// Create starts a TCP listener on the given address.
// The address format is "host:port".
func (l *TcpListener) Create(listenAddr string) error {
	addr, err := net.ResolveTCPAddr("tcp", listenAddr)
	if err != nil {
		return err
	}

	if l.listener, err = net.ListenTCP("tcp", addr); err != nil {
		return err
	}

	l.uri = fmt.Sprintf("tcp://%s", listenAddr)
	return nil
}

// Accept waits for and returns the next TCP connection.
// Returns a transport.ITransport or an error if the listener is closed.
func (l *TcpListener) Accept() (ITransport, error) {
	conn, err := l.listener.AcceptTCP()
	if err != nil {
		return nil, err
	}

	return &Tcp{Connection: conn, Err: nil}, nil
}

// Close closes the TCP listener, stopping it from accepting new connections.
func (l *TcpListener) Close() error {
	if l.listener == nil {
		return nil
	}

	return l.listener.Close()
}

// Ready returns true if the TCP listener has been created and is ready to accept connections.
func (l *TcpListener) Ready() bool {
	return l.listener != nil
}

// Uri returns the URI of the TCP listener in the format "tcp://host:port".
func (l *TcpListener) Uri() string {
	return l.uri
}

// Name returns the transport type name "TCP".
func (l *TcpListener) Name() string {
	return "TCP"
}
