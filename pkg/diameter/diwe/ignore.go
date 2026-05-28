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

type Ignore interface {
	Ignore() bool
}

func IsIgnorable(err error) bool {
	if err == nil {
		return true
	}

	var i Ignore
	return errors.As(err, &i)
}
