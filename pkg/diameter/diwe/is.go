//
// Project: TGDP - Traffic Generator for Diameter Protocol
// Description: Simple tool for testing and debugging the Diameter protocol
//
// Author: Alexander Kefeli <alexander.kefeli@gmail.com>
//
// File: error.go
// Description: Diameter pkg: AVP Debug, Info, Warnings, Errors
//

package diwe

import "errors"

// Is returns true if the error is of type T.
func Is[T error](err error) bool {
	_, ok := errors.AsType[T](err)
	return ok
}
