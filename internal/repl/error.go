//
// Project: TGDP - Traffic Generator for Diameter Protocol
// Description: Simple tool for testing and debugging the Diameter protocol
//
// Author: Alexander Kefeli <alexander.kefeli@gmail.com>
//
// File: errors.go
// Description: REPL: errors
//

package repl

// -- REPL mode errors
// --

type ErrUnknownCommand struct{}

func (e ErrUnknownCommand) Error() string {
	return "unknown command"
}

type ErrInvalidCommand struct{}

func (e ErrInvalidCommand) Error() string {
	return "invalid command"
}

type ErrQuit struct{}

func (e ErrQuit) Error() string {
	return "quit"
}
