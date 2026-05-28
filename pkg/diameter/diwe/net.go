//
// Project: TGDP - Traffic Generator for Diameter Protocol
// Description: Simple tool for testing and debugging the Diameter protocol
//
// Author: Alexander Kefeli <alexander.kefeli@gmail.com>
//
// File: error.go
// Description: Diameter pkg: Network Debug, Info, Warnings, Errors
//

package diwe

import "fmt"

// Info
//

type InfNoDataAvail struct {
	Peer string
}

func (i *InfNoDataAvail) Error() string {
	return fmt.Sprintf("No data available from peer '%s'", i.Peer)
}

func (i *InfNoDataAvail) Ignore() bool {
	return true
}

type InfInterrupted struct {
	Peer string
}

func (i *InfInterrupted) Error() string {
	return fmt.Sprintf("Peer '%s' receiving interrupted", i.Peer)
}

func (i *InfInterrupted) Ignore() bool {
	return true
}

// Warnings
//

type WarnGetRouteInfoFailed struct {
	Peer string
	Err  error
}

func (w *WarnGetRouteInfoFailed) Error() string {
	return fmt.Sprintf("Peer '%s' get route info failed: %v", w.Peer, w.Err)
}

// Errors
//

type ErrNoListeners struct {
}

func (e *ErrNoListeners) Error() string {
	return "No listeners created"
}

type ErrUnknownPeer struct {
	Peer string
}

func (e *ErrUnknownPeer) Error() string {
	return fmt.Sprintf("Unknown peer: '%s'", e.Peer)
}

type ErrPeerExists struct {
	Peer string
}

func (e *ErrPeerExists) Error() string {
	return fmt.Sprintf("Peer '%s' already exists", e.Peer)
}

type ErrNotConnected struct {
	Peer string
}

func (e *ErrNotConnected) Error() string {
	return fmt.Sprintf("Peer '%s' not connected", e.Peer)
}

type ErrConnect struct {
	Err  error
	Peer string
}

func (e *ErrConnect) Error() string {
	return fmt.Sprintf("Peer '%s' connect error: %s\n", e.Peer, e.Err)
}

type ErrAlreadyConnected struct {
	Peer string
}

func (e *ErrAlreadyConnected) Error() string {
	return fmt.Sprintf("Peer '%s' already connected", e.Peer)
}

type ErrDisconnect struct {
	Peer string
	Err  error
}

func (e *ErrDisconnect) Error() string {
	return fmt.Sprintf("Peer '%s' disconnect error: %s\n", e.Peer, e.Err)
}

type ErrSendTo struct {
	Peer string
	Err  error
}

func (e *ErrSendTo) Error() string {
	return fmt.Sprintf("Peer '%s' send error: %s", e.Peer, e.Err)
}

type ErrRecvFrom struct {
	Peer string
	Err  error
}

func (e *ErrRecvFrom) Error() string {
	return fmt.Sprintf("Peer '%s' receive error: %s", e.Peer, e.Err)
}

type ErrNoData struct {
	Peer string
}

func (e *ErrNoData) Error() string {
	return fmt.Sprintf("No data from peer '%s'", e.Peer)
}

type ErrNoSuitableAddr struct {
	Peer string
	Addr string
}

func (e *ErrNoSuitableAddr) Error() string {
	return fmt.Sprintf("Node '%s' No sutable IP address found: %s", e.Peer, e.Addr)
}
