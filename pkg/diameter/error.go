//
// Project: TGDP - Traffic Generator for Diameter Protocol
// Description: Simple tool for testing and debugging the Diameter protocol
//
// Author: Alexander Kefeli <alexander.kefeli@gmail.com>
//
// File: error.go
// Description: Diameter pkg: errors
//

package diameter

import "fmt"

// -- Application errors
// --
type ErrUnknownApp struct {
	AppId any
}

func (e *ErrUnknownApp) Error() string {
	return fmt.Sprintf("Unknown Application: '%v'", e.AppId)
}

// -- Command errors
// --
type ErrUnknownCmd struct {
	App   string
	CmdId any
}

func (e *ErrUnknownCmd) Error() string {
	return fmt.Sprintf("Unknown Command for app %s: '%v'", e.App, e.CmdId)
}

// -- AVP errors
// --
type ErrUnknownAvp struct {
	AvpId any
}

func (e *ErrUnknownAvp) Error() string {
	return fmt.Sprintf("Unknown AVP: '%v'", e.AvpId)
}

type ErrUnknownAvpType struct {
	Avp *Avp
}

func (e *ErrUnknownAvpType) Error() string {
	return fmt.Sprintf("AVP %s: unknown type '%v'", e.Avp.Name, e.Avp.Type)
}

type ErrReqAvpAbsent struct {
	AvpId any
}

func (e *ErrReqAvpAbsent) Error() string {
	return fmt.Sprintf("Mandatory AVP '%v' is absent", e.AvpId)
}

type ErrNoValueForReqAvp struct {
	AvpId any
}

func (e *ErrNoValueForReqAvp) Error() string {
	return fmt.Sprintf("Missing value for the required AVP: '%v'", e.AvpId)
}

type ErrInvalidValue struct {
	Value any
}

func (e *ErrInvalidValue) Error() string {
	return fmt.Sprintf("Invalid value '%v'", e.Value)
}

type ErrInvalidAvpValue struct {
	Avp   *Avp
	Value any
}

func (e *ErrInvalidAvpValue) Error() string {
	return fmt.Sprintf("AVP %s: invalid value '%v'", e.Avp.Name, e.Value)
}

type ErrInvalidYamlValue struct {
	Line  int
	Value any
}

func (e *ErrInvalidYamlValue) Error() string {
	return fmt.Sprintf("Invalid YAML value '%v' at line %d", e.Value, e.Line)
}

type ErrUnknownEnumItem struct {
	Avp   *Avp
	Value any
}

func (e *ErrUnknownEnumItem) Error() string {
	return fmt.Sprintf("AVP %s: unknown enum item '%v'", e.Avp.Name, e.Value)
}

type ErrAvpIsNotGrouped struct {
	Avp *Avp
}

func (e *ErrAvpIsNotGrouped) Error() string {
	return fmt.Sprintf("AVP %s is not grouped type", e.Avp.Name)
}

type ErrIndexOutOfRange struct {
	Avp   *Avp
	Index int
}

func (e *ErrIndexOutOfRange) Error() string {
	return fmt.Sprintf("Index out of data range: %d", e.Index)
}

type ErrMsgTooShort struct {
	Len int
}

func (e *ErrMsgTooShort) Error() string {
	return fmt.Sprintf("Message too short: %d", e.Len)
}
