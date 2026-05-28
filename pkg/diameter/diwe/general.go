//
// Project: TGDP - Traffic Generator for Diameter Protocol
// Description: Simple tool for testing and debugging the Diameter protocol
//
// Author: Alexander Kefeli <alexander.kefeli@gmail.com>
//
// File: error.go
// Description: Diameter pkg: General Debug, Info, Warnings, Errors
//

package diwe

// Errors
//

type ErrGeneral struct {
}

func (e *ErrGeneral) Error() string {
	return ("General error")
}

type ErrNotImplemented struct {
}

func (e *ErrNotImplemented) Error() string {
	return "Feature not implemented yet"
}

type ErrInvalidParam struct {
}

func (e *ErrInvalidParam) Error() string {
	return ("Invalid parameter")
}
