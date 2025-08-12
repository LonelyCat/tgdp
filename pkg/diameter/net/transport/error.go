//
// Project: TGDP - Traffic Generator for Diameter Protocol
// Description: Simple tool for testing and debugging the Diameter protocol
//
// Author: Alexander Kefeli <alexander.kefeli@gmail.com>
//
// File: error.go
// Description: Diameter pkg: net transport errors
//

package transport

import "fmt"

// -- Transport errors
// --
type ErrProtoUnsupported struct {
	Proto string
	Os    string
}

func (e *ErrProtoUnsupported) Error() string {
	return fmt.Sprintf("Protocol '%s' not supported on OS '%s'", e.Proto, e.Os)
}

type ErrUnknownProto struct {
	Proto string
}

func (e *ErrUnknownProto) Error() string {
	return fmt.Sprintf("Unknown protocol: %s", e.Proto)
}
