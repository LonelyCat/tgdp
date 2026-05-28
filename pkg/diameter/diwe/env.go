//
// Project: TGDP - Traffic Generator for Diameter Protocol
// Description: Simple tool for testing and debugging the Diameter protocol
//
// Author: Alexander Kefeli <alexander.kefeli@gmail.com>
//
// File: error.go
// Description: Diameter pkg: Diameter Debug, Info, Warnings, Errors
//

package diwe

import "fmt"

type ErrDiameter struct {
	Code   uint32
	CodeEx uint32
	// FIXME: Add more details here
}

func (e *ErrDiameter) Error() string {
	if e.CodeEx != 0 {
		return fmt.Sprintf("Diameter experimental result code: %d", e.CodeEx)
	}
	return fmt.Sprintf("Diameter result code: %d", e.Code)
}

type ErrInvalidMode struct {
	Mode int32
}

func (e *ErrInvalidMode) Error() string {
	return fmt.Sprintf("Invalid mode: %d", e.Mode)
}
