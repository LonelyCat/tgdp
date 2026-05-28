//
// Project: TGDP - Traffic Generator for Diameter Protocol
// Description: Simple tool for testing and debugging the Diameter protocol
//
// Author: Alexander Kefeli <alexander.kefeli@gmail.com>
//
// File: interfaces.go
// Description: Diameter pkg: interfaces between the Diameter environment and the consumers.
//

package api

// Consts
//
// Diameter Common Messages (0) command codes (RFC 6733)
const (
	AppIdCommonMessages = uint32(0)

	CmdCapabilitiesExchange = uint32(257) // Capabilities-Exchange
	CmdDeviceWatchdog       = uint32(280) // Device-Watchdog
	CmdDisconnectPeer       = uint32(282) // Disconnect-Peer-Notification
)

// Diameter result codes (RFC 6733)
const (
	DiameterSuccess = uint32(2001)
)

// Types
//

// IDiameter is the interface for the Diameter environment.
type IDiameter interface {
	CreateMessage(uint32, uint32, bool) ([]byte, error)
	CreateResponse([]byte) ([]byte, error)
	MessageHeader([]byte) (byte, uint32, uint32, uint32, byte, uint32, uint32, error)
	IsCommonMessage(uint32) bool
	IsRequest(byte) bool
	GetResultCode([]byte) (uint32, error)
	GetResultCodeEx([]byte) (uint32, error)
	TraceMessage([]byte)
}
