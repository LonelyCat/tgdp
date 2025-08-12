//
// Project: TGDP - Traffic Generator for Diameter Protocol
// Description: Simple tool for testing and debugging the Diameter protocol
//
// Author: Alexander Kefeli <alexander.kefeli@gmail.com>
//
// File: app.go
// Description: Diameter pkg: Applications handling
//

package diameter

import (
	"iter"
)

// -- Functions
// --
func (a *App) CmdIter() iter.Seq[*Cmd] {
	return func(yield func(*Cmd) bool) {
		for _, value := range a.Cmds {
			if !yield(value) {
				break
			}
		}
	}
}

func (a *App) CmdIter2() iter.Seq2[int, *Cmd] {
	return func(yield func(int, *Cmd) bool) {
		for idx, value := range a.Cmds {
			if !yield(idx, value) {
				break
			}
		}
	}
}
