//
// Project: TGDP - Traffic Generator for Diameter Protocol
// Description: Simple tool for testing and debugging the Diameter protocol
//
// Author: Alexander Kefeli <alexander.kefeli@gmail.com>
//
// File: error.go
// Description: Diameter pkg: Message Debug, Info, Warnings, Errors
//

package diwe

import "fmt"


type ErrMsgTooShort struct {
	Len int
}

func (e *ErrMsgTooShort) Error() string {
	return fmt.Sprintf("Message too short: %d", e.Len)
}

type ErrMissingAvp struct {
	Avp any
}

func (e *ErrMissingAvp) Error() string {
	return fmt.Sprintf("AVP is missing in message: '%v'", e.Avp)
}

type ErrMissingReqAvp struct {
	Avp any
}

func (e *ErrMissingReqAvp) Error() string {
	return fmt.Sprintf("Mandatory AVP is missing: '%v'", e.Avp)
}

type ErrNoValueForReqAvp struct {
	Avp any
}

func (e *ErrNoValueForReqAvp) Error() string {
	return fmt.Sprintf("Missing value for the mandatory AVP: '%v'", e.Avp)
}

type ErrNoAvpValue struct {
	Avp any
}

func (e *ErrNoAvpValue) Error() string {
	return fmt.Sprintf("No value found for AVP '%v'", e.Avp)
}
