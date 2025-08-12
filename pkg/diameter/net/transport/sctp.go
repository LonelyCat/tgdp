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
	"fmt"
	"net"
	"syscall"
	"time"

	"github.com/ishidawataru/sctp"
)

// -- Types
// --
type Sctp struct {
	Connection *sctp.SCTPConn
	Err        error
}

// -- Methods
// --
func (t *Sctp) Connect(remoteAddr net.IP, remotePort int, localAddr net.IP, localPort int) error {
	rAddr := &sctp.SCTPAddr{
		IPAddrs: []net.IPAddr{
			{IP: remoteAddr},
		},
		Port: remotePort,
	}

	lAddr := &sctp.SCTPAddr{
		IPAddrs: []net.IPAddr{
			{IP: localAddr},
		},
		Port: localPort,
	}

	for {
		if conn, err := sctp.DialSCTP("sctp", lAddr, rAddr); err != nil {
			switch err.(syscall.Errno) {
			case syscall.EISCONN, syscall.EALREADY, syscall.EADDRINUSE:
				fmt.Println("Connection already established")
				fmt.Println("Closing connection")
				fmt.Println("Trying again")
				_ = conn.Close()
				time.Sleep(time.Second * 1)
				continue
			default:
				t.Err = err
				return err
			}
		} else {
			t.Connection = conn
			break
		}
	}

	return nil
}

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
	return t.Connection.RemoteAddr().String()
}

func (t *Sctp) LocalAddr() string {
	return t.Connection.LocalAddr().String()
}

func (t *Sctp) RemoteIp() net.IP {
	return t.Connection.RemoteAddr().(*sctp.SCTPAddr).IPAddrs[0].IP
}

func (t *Sctp) LocalIp() net.IP {
	return t.Connection.LocalAddr().(*sctp.SCTPAddr).IPAddrs[0].IP
}

func (t *Sctp) RemotePort() int {
	return t.Connection.RemoteAddr().(*sctp.SCTPAddr).Port
}

func (t *Sctp) LocalPort() int {
	return t.Connection.LocalAddr().(*sctp.SCTPAddr).Port
}

func (t *Sctp) Error() error {
	return t.Err
}

func (t *Sctp) Name() string {
	return "SCTP"
}
