//
// Project: TGDP - Traffic Generator for Diameter Protocol
// Description: Simple tool for testing and debugging the Diameter protocol
//
// Author: Alexander Kefeli <alexander.kefeli@gmail.com>
//
// File: error.go
// Description: Diameter pkg: net node errors
//

package node

import "fmt"

// -- Peer errors
// --
type ErrReadYaml struct {
	File string
	Err  error
}

func (e *ErrReadYaml) Error() string {
	return fmt.Sprintf("Failed %v", e.Err)
}

type ErrParseYaml struct {
	File string
	Err  error
}

func (e *ErrParseYaml) Error() string {
	return fmt.Sprintf("Failed to parse '%s': %v", e.File, e.Err)
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

type ErrAlreadyConnected struct {
	Peer string
}

func (e *ErrAlreadyConnected) Error() string {
	return fmt.Sprintf("Peer '%s' already connected", e.Peer)
}

func (e *ErrConnect) Error() string {
	return fmt.Sprintf("Peer '%s' connect error: %s\n", e.Peer, e.Err)
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
	return fmt.Sprintf("Peer '%s' recv error: %s", e.Peer, e.Err)
}

type ErrNoData struct {
	Peer string
}

func (e *ErrNoData) Error() string {
	return fmt.Sprintf("No data from peer '%s'", e.Peer)
}

type ErrNoSuitableAddr struct {
	Addr string
}

func (e *ErrNoSuitableAddr) Error() string {
	return fmt.Sprintf("No sutable IP address found: %s", e.Addr)
}

type ErrInterrupted struct {
}

func (e *ErrInterrupted) Error() string {
	return "Interrupted"
}

type ErrDiameter struct {
	Code uint32
	// FIXME: Add more details here
}

func (e *ErrDiameter) Error() string {
	return fmt.Sprintf("Diameter error: %d", e.Code)
}
