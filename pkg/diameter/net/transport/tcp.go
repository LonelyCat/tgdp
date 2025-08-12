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
	"net"
	"time"
)

// -- Types
// --
type Tcp struct {
	Connection *net.TCPConn
	Err        error
}

// -- Methods
// --
func (t *Tcp) Connect(remoteAddr net.IP, remotePort int, localAddr net.IP, localPort int) error {
	rAddr := &net.TCPAddr{
		IP: remoteAddr,
		Port: remotePort,
	}

	lAddr := &net.TCPAddr{
		IP: localAddr,
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
		// t.Connection = nil
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
	return (*t.Connection).RemoteAddr().String()
}

func (t *Tcp) LocalAddr() string {
	return (*t.Connection).LocalAddr().String()
}

func (t *Tcp) RemoteIp() net.IP {
	return net.ParseIP(t.Connection.RemoteAddr().String())
}

func (t *Tcp) LocalIp() net.IP {
	return net.ParseIP(t.Connection.LocalAddr().String())
}

func (t *Tcp) RemotePort() int {
	return int(t.Connection.RemoteAddr().(*net.TCPAddr).IP[0])
}

func (t *Tcp) LocalPort() int {
	return int(t.Connection.LocalAddr().(*net.TCPAddr).IP[0])
}

func (t *Tcp) Error() error {
	return t.Err
}

func (t *Tcp) Name() string {
	return "TCP"
}
