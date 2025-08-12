//
// Project: TGDP - Traffic Generator for Diameter Protocol
// Description: Simple tool for testing and debugging the Diameter protocol
//
// Author: Alexander Kefeli <alexander.kefeli@gmail.com>
//
// File: message.go
// Description: Diameter pkg: message handling
//

package diameter

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log/slog"
	"math/rand"
	"slices"
	"strings"
	"time"
)

// -- Consts
// --
const (
	MinMessageLen = uint32(20)
)

// -- Types
// --
type Message struct {
	Version  uint8
	Length   uint32
	AppId    uint32
	Flags    uint8
	CmdCode  uint32
	HopByHop uint32
	EndToEnd uint32
	Avps     []*Avp
	Bytes    []byte
}

// -- Functions
// --
func NewMessage(app *App, cmd *Cmd, request, fetchAvps bool) (*Message, error) {
	m := new(Message)
	m.Version = 1 // Version 1 for now
	m.AppId = app.Id
	m.CmdCode = cmd.Code
	m.Flags = cmd.Flags
	m.Length = MinMessageLen
	m.EndToEnd = rand.Uint32()
	m.HopByHop = rand.Uint32()
	m.Avps = make([]*Avp, 0)

	if request {
		m.Flags |= Dict.CmdFlagR()
	}

	if !fetchAvps {
		return m, nil
	}

	var avpRules []*AvpRule
	if request {
		avpRules = cmd.Request
	} else {
		avpRules = cmd.Answer
	}

	if size, err := FetchAvpsValues(avpRules, &m.Avps, nil); err != nil {
		return nil, err
	} else {
		m.Length += size
	}

	if sessionId, err := m.GetAvp("Session-Id"); err == nil {
		value := sessionId.GetValue().(string)
		now := time.Now().UnixNano()
		var v any = fmt.Sprintf("%s;%d;%d", value, now>>32, now%0xFFFFFFFF)
		_ = sessionId.SetValue(&v)
	}

	return m, nil
}

func NewMessage2(appId any, cmdId any, request, fetchAvps bool) (*Message, error) {
	app, err := Dict.GetApp(appId)
	if err != nil {
		return nil, err
	}

	cmd, err := Dict.GetCmd(cmdId, app)
	if err != nil {
		return nil, err
	}

	return NewMessage(app, cmd, request, fetchAvps)
}

func Request(appId uint32, cmdCode uint32) (*Message, error) {
	return NewMessage2(appId, cmdCode, true, true)
}

func Answer(appId uint32, cmdCode uint32) (*Message, error) {
	return NewMessage2(appId, cmdCode, false, true)
}

// -- Methods
// --
func (m *Message) Reply() (*Message, error) {

	r, err := NewMessage2(m.AppId, m.CmdCode, false, true)
	if err != nil {
		return nil, err
	}

	if sessionId, err := m.GetAvp("Session-Id"); err == nil {
		var value any = sessionId.GetValue().(string)
		if sessionId, err := r.GetAvp("Session-Id"); err == nil {
			_ = sessionId.SetValue(&value)
		}
	}

	r.HopByHop = m.HopByHop
	r.EndToEnd = m.EndToEnd
	r.Flags = m.Flags & ^Dict.CmdFlagR()

	return r, err
}

func (m *Message) AddAvp(avpId any) error {
	var avp *Avp
	var err error
	switch avpId := avpId.(type) {
	case uint32:
		avp, err = Dict.GetAvpByCode(avpId)
	case string:
		avp, err = Dict.GetAvpByName(avpId)
	case *Avp:
		avp = avpId
	}
	if err != nil {
		return err
	}
	// FIXME: check for max AVP count
	m.Avps = append(m.Avps, avp)
	m.Bytes = nil

	return nil
}

func (m *Message) RemoveAvp(avpId any) error {
	for i, avp := range m.Avps {
		if matchAvp(avp, avpId) {
			m.Avps = slices.Delete(m.Avps, i, i+1)
			m.Bytes = nil
			return nil
		}
	}

	return &ErrUnknownAvp{AvpId: avpId}
}

func (m *Message) GetAvp(avpId any) (*Avp, error) {
	for _, avp := range m.Avps {
		if matchAvp(avp, avpId) {
			return avp, nil
		}
	}

	return nil, &ErrUnknownAvp{AvpId: avpId}
}

func (m *Message) GetAvp2(avpId any, index int) (*Avp, error) {
	n := 1
	for _, avp := range m.Avps {
		if matchAvp(avp, avpId) {
			if n == index {
				return avp, nil
			}
			n++
		}
	}

	return nil, &ErrUnknownAvp{AvpId: avpId}
}

func (m *Message) Serialize() ([]byte, error) {
	if m.Bytes != nil {
		return m.Bytes, nil
	}

	buf := new(bytes.Buffer)

	binary.Write(buf, binary.BigEndian, uint32(m.Version)<<24|uint32(m.Length&0x00FFFFFF))
	binary.Write(buf, binary.BigEndian, uint32(m.Flags)<<24|uint32(m.CmdCode&0x00FFFFFF))
	binary.Write(buf, binary.BigEndian, m.AppId)
	binary.Write(buf, binary.BigEndian, m.HopByHop)
	binary.Write(buf, binary.BigEndian, m.EndToEnd)

	for _, avp := range m.Avps {
		if err := avp.Serialize(buf); err != nil {
			return nil, err
		}
	}

	m.Length = uint32(buf.Len())
	binary.BigEndian.PutUint32(buf.Bytes(), (1<<24)|(m.Length)&0x00FFFFFF)
	m.Bytes = buf.Bytes()
	return m.Bytes, nil
}

func (m *Message) Deserialize(data []byte) error {
	if len(data) < int(MinMessageLen) {
		return &ErrMsgTooShort{len(data)}
	}

	m.Version = uint8(data[0])
	m.Length = binary.BigEndian.Uint32(data[0:4]) & 0x00FFFFFF
	m.Flags = uint8(data[4])
	m.CmdCode = binary.BigEndian.Uint32(data[4:8]) & 0x00FFFFFF
	m.AppId = binary.BigEndian.Uint32(data[8:12])
	m.HopByHop = binary.BigEndian.Uint32(data[12:16])
	m.EndToEnd = binary.BigEndian.Uint32(data[16:20])

	offset := uint32(MinMessageLen)
	for offset < uint32(len(data)) {
		avp := new(Avp)
		o, err := avp.Deserialize(data[offset:])
		if err != nil {
			slog.Warn(err.Error())
		}
		offset += o
		m.Avps = append(m.Avps, avp)
	}
	m.Bytes = make([]byte, len(data))
	copy(m.Bytes, data)

	return nil
}

func (m *Message) IsRequest() bool {
	return (m.Flags&Dict.CmdFlagR() != 0)
}

func (m *Message) IsProxyable() bool {
	return (m.Flags&Dict.CmdFlagP() != 0)
}

func (m *Message) IsError() bool {
	return (m.Flags&Dict.CmdFlagE() != 0)
}

func (m *Message) IsRetransmition() bool {
	return (m.Flags&Dict.CmdFlagT() != 0)
}

func (m *Message) Buff() []byte {
	return m.Bytes
}

func (m *Message) Len() uint32 {
	return m.Length
}

func (m *Message) Dump(shift ...int) {
	app, err := Dict.GetAppById(m.AppId)
	if app == nil {
		slog.Error(err.Error())
		return
	}

	cmd, err := Dict.GetCmdByCode(m.CmdCode, app)
	if cmd == nil {
		slog.Error(err.Error())
		return
	}

	fmt.Printf("Version:  %d\n", m.Version)
	fmt.Printf("Length:   %d\n", m.Length)
	fmt.Printf("AppId:    %d (%s)\n", m.AppId, app.Name)
	fmt.Printf("Flags:    0x%02X", m.Flags)
	if m.Flags != 0 {
		fmt.Print(" (")
		flags := []string{}
		for flag := Dict.CmdFlagR(); flag > Dict.CmdFlagT(); flag >>= 1 {
			if flag&m.Flags != 0 {
				flags = append(flags, Dict.CmdFlags.String(flag))
			}
		}
		fmt.Print(strings.Join(flags, ", "))
		fmt.Print(")")
	}
	fmt.Println()
	fmt.Printf("CmdCode:  %d (%s)\n", m.CmdCode, cmd.Name)
	fmt.Printf("EndToEnd: 0x%08X (%d)\n", m.EndToEnd, m.EndToEnd)
	fmt.Printf("HopByHop: 0x%08X (%d)\n", m.HopByHop, m.HopByHop)
	fmt.Printf("AVPs:\n")
	for _, avp := range m.Avps {
		avp.Dump(2)
	}
	fmt.Println()
}

// -- Helpers
// --
func matchAvp(avp *Avp, avpId any) bool {
	switch avpId := avpId.(type) {
	case int:
		return avp.Code == uint32(avpId)
	case uint32:
		return avp.Code == avpId
	case string:
		return strings.EqualFold(avp.Name, avpId)
	case *Avp:
		return avp == avpId
	default:
		return false
	}
}
