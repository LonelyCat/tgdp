//
// Project: TGDP - Traffic Generator for Diameter Protocol
// Description: Simple tool for testing and debugging the Diameter protocol
//
// Author: Alexander Kefeli <alexander.kefeli@gmail.com>
//
// File: error.go
// Description: Diameter pkg: PCAP files errors
//

package pcap

import "fmt"

// -- PCAP errors
// --
type ErrOpenFile struct {
	File string
	Err  error
}

func (e *ErrOpenFile) Error() string {
	return fmt.Sprintf("Error open file %s: %v", e.File, e.Err)
}

type ErrWriteFile struct {
	File string
	Err  error
}

func (e *ErrWriteFile) Error() string {
	return fmt.Sprintf("Error write file %s: %v", e.File, e.Err)
}

type ErrSerLayers struct {
	Err error
}

func (e *ErrSerLayers) Error() string {
	return fmt.Sprintf("Failed to serialize PCAP layers: %v\n", e.Err)
}
