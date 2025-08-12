//
// Project: TGDP - Traffic Generator for Diameter Protocol
// Description: Simple tool for testing and debugging the Diameter protocol
//
// Author: Alexander Kefeli <alexander.kefeli@gmail.com>
//
// File: diameter.go
// Description: Diameter pkg: dictionary handling
//

package diameter

import (
	"context"
	"fmt"
	"iter"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"tgdp/internal/config"
)

// -- Variables
// --
var (
	Dict *Diameter
)

// -- Methods
// --
func (d *Diameter) GetApp(appId any) (*App, error) {
	switch appId := appId.(type) {
	case uint32:
		return d.GetAppById(appId)
	case string:
		if id, err := strconv.ParseUint(appId, 10, 32); err == nil {
			return d.GetAppById(uint32(id))
		}
		return d.GetAppByName(appId)
	default:
		return nil, &ErrUnknownApp{AppId: appId}
	}
}

func (d *Diameter) GetAppById(appId uint32) (*App, error) {
	for _, app := range Dict.Apps {
		if app.Id == appId {
			return app, nil
		}
	}

	return nil, &ErrUnknownApp{AppId: appId}
}

func (d *Diameter) GetAppByName(appName string) (*App, error) {
	for _, app := range Dict.Apps {
		if strings.EqualFold(app.Name, appName) {
			return app, nil
		}
	}

	return nil, &ErrUnknownApp{AppId: appName}
}

func (d *Diameter) GetCmd(cmdId any, app *App) (*Cmd, error) {
	switch cmdId := cmdId.(type) {
	case uint32:
		return d.GetCmdByCode(cmdId, app)
	case string:
		if code, err := strconv.ParseUint(cmdId, 10, 32); err == nil {
			return d.GetCmdByCode(uint32(code), app)
		}
		return d.GetCmdByName(cmdId, app)
	default:
		return nil, &ErrUnknownCmd{App: app.Name, CmdId: cmdId}
	}
}

func (d *Diameter) GetCmdByCode(code uint32, app *App) (*Cmd, error) {
	for _, cmd := range app.Cmds {
		if cmd.Code == code {
			return cmd, nil
		}
	}

	return nil, &ErrUnknownCmd{App: app.Name, CmdId: code}
}

func (d *Diameter) GetCmdByName(cmdName string, app *App) (*Cmd, error) {
	for _, cmd := range app.Cmds {
		if strings.EqualFold(cmd.Short, cmdName) {
			return cmd, nil
		}
	}

	return nil, &ErrUnknownCmd{App: app.Name, CmdId: cmdName}
}

func (d *Diameter) GetAvp(avpId any) (*Avp, error) {
	switch avpId := avpId.(type) {
	case uint32:
		return d.GetAvpByCode(avpId)
	case string:
		if code, err := strconv.ParseUint(avpId, 10, 32); err == nil {
			return d.GetAvpByCode(uint32(code))
		}
		return d.GetAvpByName(avpId)
	default:
		return nil, &ErrUnknownAvp{AvpId: avpId}
	}
}

func (d *Diameter) GetAvpByCode(avpCode uint32) (*Avp, error) {
	for _, avp := range d.Avps {
		if avp.Code == avpCode {
			return avp.clone(), nil
		}
	}
	return nil, &ErrUnknownAvp{AvpId: avpCode}
}

func (d *Diameter) GetAvpByName(avpName string) (*Avp, error) {
	for _, avp := range d.Avps {
		if strings.EqualFold(avp.Name, avpName) {
			return avp.clone(), nil
		}
	}
	return nil, &ErrUnknownAvp{AvpId: avpName}
}

func (d *Diameter) AvpFlagV() uint8 {
	return d.AvpFlags.V
}

func (d *Diameter) AvpFlagM() uint8 {
	return d.AvpFlags.M
}

func (d *Diameter) AvpFlagP() uint8 {
	return d.AvpFlags.P
}

func (d *Diameter) CmdFlagR() uint8 {
	return d.CmdFlags.R
}

func (d *Diameter) CmdFlagP() uint8 {
	return d.CmdFlags.P
}

func (d *Diameter) CmdFlagE() uint8 {
	return d.CmdFlags.E
}

func (d *Diameter) CmdFlagT() uint8 {
	return d.CmdFlags.T
}

func (d *Diameter) AvpTypeOctetString() int {
	return d.AvpTypes.OctetString
}

func (d *Diameter) AvpTypeInteger32() int {
	return d.AvpTypes.Integer32
}

func (d *Diameter) AvpTypeInteger64() int {
	return d.AvpTypes.Integer64
}

func (d *Diameter) AvpTypeUnsigned32() int {
	return d.AvpTypes.Unsigned32
}

func (d *Diameter) AvpTypeUnsigned64() int {
	return d.AvpTypes.Unsigned64
}

func (d *Diameter) AvpTypeFloat32() int {
	return d.AvpTypes.Float32
}

func (d *Diameter) AvpTypeFloat64() int {
	return d.AvpTypes.Float64
}

func (d *Diameter) AvpTypeAddress() int {
	return d.AvpTypes.Address
}

func (d *Diameter) AvpTypeTime() int {
	return d.AvpTypes.Time
}

func (d *Diameter) AvpTypeUTF8String() int {
	return d.AvpTypes.UTF8String
}

func (d *Diameter) AvpTypeIdentity() int {
	return d.AvpTypes.Identity
}

func (d *Diameter) AvpTypeURI() int {
	return d.AvpTypes.URI
}

func (d *Diameter) AvpTypeIPFilterRule() int {
	return d.AvpTypes.IPFilterRule
}

func (d *Diameter) AvpTypeQoSFilterRule() int {
	return d.AvpTypes.QoSFilterRule
}

func (d *Diameter) AvpTypeEnumerated() int {
	return d.AvpTypes.Enumerated
}

func (d *Diameter) AvpTypeGrouped() int {
	return d.AvpTypes.Grouped
}

func (d *Diameter) AppIter() iter.Seq[*App] {
	return func(yield func(*App) bool) {
		for _, value := range d.Apps {
			if !yield(value) {
				break
			}
		}
	}
}

func (d *Diameter) AppIter2() iter.Seq2[int, *App] {
	return func(yield func(int, *App) bool) {
		for idx, value := range d.Apps {
			if !yield(idx, value) {
				break
			}
		}
	}
}

func (d *Diameter) AvpIter() iter.Seq[*Avp] {
	return func(yield func(*Avp) bool) {
		for _, value := range d.Avps {
			if !yield(value) {
				break
			}
		}
	}
}

func (d *Diameter) AvpIter2() iter.Seq2[int, *Avp] {
	return func(yield func(int, *Avp) bool) {
		for idx, value := range d.Avps {
			if !yield(idx, value) {
				break
			}
		}
	}
}

func (d *Diameter) Show() {
	for _, app := range d.Apps {
		fmt.Printf("Application: %s (%d)\n", app.Name, app.Id)
		for _, cmd := range app.Cmds {
			fmt.Printf("  %s: %s (%d)\n", cmd.Short, cmd.Name, cmd.Code)
		}
		fmt.Println()
	}
}

func (d *Diameter) Verify() int {
	errFound := 0

	for _, app := range d.Apps {
		for _, cmd := range app.Cmds {
			for _, rule := range cmd.Request {
				if avp, _ := d.GetAvpByName(rule.Name); avp == nil {
					fmt.Printf("%s/%sR: unknown AVP \"%s\"\n", app.Name, cmd.Short, rule.Name)
					errFound++
				}
			}
			for _, rule := range cmd.Answer {
				if avp, _ := d.GetAvpByName(rule.Name); avp == nil {
					fmt.Printf("%s/%sA: unknown AVP \"%s\"\n", app.Name, cmd.Short, rule.Name)
					errFound++
				}
			}
		}
	}
	fmt.Println()

	avpCodes := make(map[uint32]uint32)
	for _, avp := range Dict.Avps {
		if code, exists := avpCodes[avp.Code]; exists {
			avpDup, _ := Dict.GetAvpByCode(code)
			fmt.Printf("Duplicated code %d for AVPs \"%s\" and \"%s\"\n", code, avp.Name, avpDup.Name)
			errFound++
		} else {
			avpCodes[avp.Code] = avp.Code
		}
	}
	fmt.Println()

	for _, avp := range Dict.Avps {
		if avp.Flags&d.AvpFlagV() != 0 && avp.VndId == 0 {
			fmt.Printf("AVP \"%s\" V-flag persent without vendor id\n", avp.Name)
			errFound++
		}
	}
	fmt.Println()

	for _, avp := range Dict.Avps {
		if avp.Type == d.AvpTypeGrouped() {
			for _, member := range avp.Group.Members {
				memberAvp, _ := d.GetAvpByName(member.Name)
				if memberAvp == nil {
					fmt.Printf("AVP \"%s\" unknown group member: \"%s\"\n", avp.Name, member.Name)
					errFound++
				}
			}
		}
	}
	fmt.Println()

	if errFound > 0 {
		fmt.Println(">>> Errors: ", errFound)
	} else {
		fmt.Println(">>> No errors foud :)")
	}
	return errFound
}

func (d *Diameter) Dump(n ...int) {
	fmt.Println("\n>>> Applications")
	for _, app := range d.Apps {
		fmt.Printf("id=%d \"%s\"\n", app.Id, app.Name)
		for _, cmd := range app.Cmds {
			fmt.Printf("\tcode=%d \"%s\" \"%s\"\n", cmd.Code, cmd.Short, cmd.Name)
			fmt.Println("\t\tRequest")
			for _, avp := range cmd.Request {
				fmt.Printf("\t\t\t\"%s\"\n", avp.Name)
			}
			fmt.Println("\t\tAnswer")
			for _, avp := range cmd.Answer {
				fmt.Printf("\t\t\t\"%s\"\n", avp.Name)
			}
		}
	}

	fmt.Println("\n>>> AVPs")
	for _, avp := range d.Avps {
		fmt.Printf("code=%d name=\"%s\" flags=%d type=%d\n", avp.Code, avp.Name, avp.Flags, avp.Type)
		if avp.Type == 13 {
			for _, item := range avp.Enum.Items {
				fmt.Printf("\tcode=%d name=\"%s\"\n", item.Code, item.Name)
			}
		}
		if avp.Type == 16 {
			for _, member := range avp.Group.Members {
				fmt.Printf("\tname=\"%s\"\n", member.Name)
			}
		}
	}
}

// -- Functions
// --
func LoadDictionary() error {
	path := filepath.Join(config.ConfDir(), config.DiaDictPkl)
	d, err := LoadFromPath(context.Background(), path)
	if err != nil {
		return err
	}
	Dict = (*Diameter)(d)

	dia2Go = map[int]avpGoType{
		Dict.AvpTypeAddress():       {reflect.TypeOf(""), reflect.TypeOf("")},
		Dict.AvpTypeEnumerated():    {reflect.TypeOf(""), reflect.TypeOf(0)},
		Dict.AvpTypeIdentity():      {reflect.TypeOf(""), reflect.TypeOf("")},
		Dict.AvpTypeOctetString():   {reflect.TypeOf(""), reflect.TypeOf(0)},
		Dict.AvpTypeIPFilterRule():  {reflect.TypeOf(""), reflect.TypeOf("")},
		Dict.AvpTypeQoSFilterRule(): {reflect.TypeOf(""), reflect.TypeOf("")},
		Dict.AvpTypeTime():          {reflect.TypeOf(""), reflect.TypeOf("")},
		Dict.AvpTypeUTF8String():    {reflect.TypeOf(""), reflect.TypeOf("")},
		Dict.AvpTypeURI():           {reflect.TypeOf(""), reflect.TypeOf("")},
		Dict.AvpTypeInteger32():     {reflect.TypeOf(int32(0)), reflect.TypeOf(0)},
		Dict.AvpTypeInteger64():     {reflect.TypeOf(int64(0)), reflect.TypeOf(0)},
		Dict.AvpTypeUnsigned32():    {reflect.TypeOf(uint32(0)), reflect.TypeOf(0)},
		Dict.AvpTypeUnsigned64():    {reflect.TypeOf(uint64(0)), reflect.TypeOf(0)},
		Dict.AvpTypeFloat32():       {reflect.TypeOf(float32(0.0)), reflect.TypeOf(0)},
		Dict.AvpTypeFloat64():       {reflect.TypeOf(float64(0.0)), reflect.TypeOf(0)},
		Dict.AvpTypeGrouped():       {reflect.TypeOf(&AvpDataStore{}), nil},
	}

	return LoadAvpsDataFromYaml(filepath.Join(config.ConfDir(), config.AvpDataFile), AVP_DATA_APPEND)
}

func (f *CmdBitFlags) String(flag uint8) string {
	switch flag {
	case Dict.CmdFlagR():
		return "Request"
	case Dict.CmdFlagP():
		return "Proxyable"
	case Dict.CmdFlagE():
		return "Error"
	case Dict.CmdFlagT():
		return "Retransmition"
	default:
		return "Unknown"
	}
}

func (f *AvpBitFlags) String(args ...any) string {
	flag := args[0].(uint8)
	switch flag {
	case Dict.AvpFlagM():
		return "M"
	case Dict.AvpFlagV():
		return "V"
	case Dict.AvpFlagP():
		return "P"
	}
	return "-"
}

func (dt *AvpDataTypes) String(id int) string {
	t := reflect.TypeOf(*dt)

	if id <= t.NumField() {
		return t.Field(id - 1).Name
	} else {
		return "unknown"
	}
}
