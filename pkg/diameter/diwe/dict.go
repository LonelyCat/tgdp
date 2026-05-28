//
// Project: TGDP - Traffic Generator for Diameter Protocol
// Description: Simple tool for testing and debugging the Diameter protocol
//
// Author: Alexander Kefeli <alexander.kefeli@gmail.com>
//
// File: error.go
// Description: Diameter pkg: Dictionary Debug, Info, Warnings, Errors
//

package diwe

import "fmt"

type ErrUnknownFmt struct {
	Fmt int
}

func (e *ErrUnknownFmt) Error() string {
	return fmt.Sprintf("Unknown format: '%v'", e.Fmt)
}

type ErrUnknownApp struct {
	AppId any
}

func (e *ErrUnknownApp) Error() string {
	return fmt.Sprintf("Unknown Application: '%v'", e.AppId)
}

type ErrUnknownCmd struct {
	App   string
	CmdId any
}

func (e *ErrUnknownCmd) Error() string {
	return fmt.Sprintf("Unknown Command for app %s: '%v'", e.App, e.CmdId)
}

type ErrUnknownAvp struct {
	Avp any
}

func (e *ErrUnknownAvp) Error() string {
	return fmt.Sprintf("Unknown AVP: '%v'", e.Avp)
}

type ErrUnknownAvpType struct {
	Avp  string
	Type int
}

func (e *ErrUnknownAvpType) Error() string {
	return fmt.Sprintf("AVP %s: unknown type '%v'", e.Avp, e.Type)
}

type ErrUnknownEnumItem struct {
	Avp   string
	Value any
}

func (e *ErrUnknownEnumItem) Error() string {
	return fmt.Sprintf("AVP %s: unknown enum item '%v'", e.Avp, e.Value)
}
