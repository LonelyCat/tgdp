//
// Project: TGDP - Traffic Generator for Diameter Protocol
// Description: Simple tool for testing and debugging the Diameter protocol
//
// Author: Alexander Kefeli <alexander.kefeli@gmail.com>
//
// File: debug.go
// Description: Diameter pkg: control debug and verbosity
//

package diameter

import (
	"fmt"
)

// -- Consts
// --
const (
	VerboseMsg = iota
	VerboseAvp
	VerbosePeer
	VerboseCM
)

// -- Types
// --
type IDebug interface {
	Dump(shift ...int)
}

type IText interface {
	String(...any) string
}

// -- Variables
// --
var verboseLevel int = 0

// -- Functions
// --
func GetVerboseLevel() int {
	return verboseLevel
}

func SetVerboseLevel(level int) {
	verboseLevel = level
}

func Verbose(obj IDebug, level int) {
	if level <= verboseLevel {
		fmt.Println()
		obj.Dump()
	}
}
