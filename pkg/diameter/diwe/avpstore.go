//
// Project: TGDP - Traffic Generator for Diameter Protocol
// Description: Simple tool for testing and debugging the Diameter protocol
//
// Author: Alexander Kefeli <alexander.kefeli@gmail.com>
//
// File: error.go
// Description: Diameter pkg: ASP store Debug, Info, Warnings, Errors
//

package diwe

import "fmt"

type ErrUnknownStoreAction struct {
	Action int
}

func (e *ErrUnknownStoreAction) Error() string {
	return fmt.Sprintf("Unknown store action: %d", e.Action)
}

type ErrIndexOutOfRange struct {
	Index int
}

func (e *ErrIndexOutOfRange) Error() string {
	return fmt.Sprintf("Index out of data range: %d", e.Index)
}
