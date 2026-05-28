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

import "fmt"

type ErrAvpTooShort struct {
	Len int
}

func (e *ErrAvpTooShort) Error() string {
	return fmt.Sprintf("AVP too short: %d", e.Len)
}

type ErrInvalidValue struct {
	Value any
}

func (e *ErrInvalidValue) Error() string {
	return fmt.Sprintf("Invalid value '%v'", e.Value)
}

type ErrInvalidAvpValue struct {
	Avp   any
	Value any
}

func (e *ErrInvalidAvpValue) Error() string {
	return fmt.Sprintf("AVP %s: invalid value '%v'", e.Avp, e.Value)
}

type ErrInvalidYamlValue struct {
	Line   int
	Column int
	Value  any
}

func (e *ErrInvalidYamlValue) Error() string {
	return fmt.Sprintf("Invalid YAML value '%v' at line %d", e.Value, e.Line)
}

type ErrAvpIsNotGrouped struct {
	AvpName string
}

func (e *ErrAvpIsNotGrouped) Error() string {
	return fmt.Sprintf("AVP %s is not grouped type", e.AvpName)
}
