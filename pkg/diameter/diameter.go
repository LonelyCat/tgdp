//
// Project: TGDP - Traffic Generator for Diameter Protocol
// Description: Simple tool for testing and debugging the Diameter protocol
//
// Author: Alexander Kefeli <alexander.kefeli@gmail.com>
//
// File: diameter.go
// Description: Diameter pkg: Diameter environment
//

package diameter

import (
	"context"
	"encoding/binary"
	"fmt"
	"log/slog"
	"math/rand"
	"os"
	"reflect"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"tgdp/pkg/diameter/api"
	"tgdp/pkg/diameter/dict"
	"tgdp/pkg/diameter/diwe"
	"tgdp/pkg/diameter/net/node"
	"tgdp/pkg/diameter/net/transport"
	"tgdp/pkg/diameter/pcap"
)

// Consts
//

// Context key for Diameter environment
type EnvContextKey string

const (
	EnvContext EnvContextKey = "diameter-env"
)

const (
	mask24bits = 0x00FFFFFF
	mask32bits = 0xFFFFFFFF
)

var (
	avpPool = sync.Pool{
		New: func() any {
			return &Avp{}
		},
	}

	messagePool = sync.Pool{
		New: func() any {
			return &Message{}
		},
	}
)

// Mode constants define the Diameter operating mode
const (
	ModeTransaction = int32(iota)
	ModeSession
)

// Trace level constants define logging verbosity
const (
	TraceQuiet = int32(iota)
	TraceMsg
	TracePeer
	TraceCM
)

// AVP codes (RFC 6733)
const (
	avpResultCode = uint32(268) // Result-Code
	avpSessionId  = uint32(263) // Session-Id
)

// Diameter version
const (
	Version = 1
)

// Types
//
// Diameter represents the Diameter protocol environment with configuration,
// dictionary, peer management, and AVP storage capabilities.
type Diameter struct {
	mode    atomic.Int32
	dict    dict.Dict
	peers   node.Nodes
	store   AvpStore
	codecs  AvpCodecs
	verbLvl atomic.Int32
	dia2go  diaTypesToGo
	ctx     context.Context
	cancel  context.CancelFunc
	wgDone  sync.WaitGroup
	pcap    *pcap.Pcap
	logger  *slog.Logger
	rng     *rand.Rand
}

// diaTypesToGo maps Diameter data types to Go types for serialization/deserialization
type diaTypesToGo map[int]avpGoType

// avpGoType holds reflect.Type information for AVP value conversions
type avpGoType struct {
	type1 reflect.Type
	type2 reflect.Type
}

// IDebug defines the interface for debug output functionality
type ITrace interface {
	Trace(shift ...int)
}

// Constructor
//
// New creates and initializes a new Diameter environment with the specified mode.
// Returns an error if the mode is invalid.
func New(mode int32) (*Diameter, error) {
	source := rand.NewSource(time.Now().UnixNano())

	logOptions := &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}
	handler := slog.NewTextHandler(os.Stdout, logOptions)
	logger := slog.New(handler)

	d := &Diameter{
		peers:  node.NewNodes(),
		codecs: make(AvpCodecs),
		dia2go: make(diaTypesToGo),
		rng:    rand.New(source),
		logger: logger,
	}
	d.store = NewAvpStore(d)
	d.ctx, d.cancel = context.WithCancel(context.Background())
	return d, d.SetMode(mode)
}

// Methods
//
// Context returns the context for the Diameter environment context.
func (d *Diameter) Context() context.Context {
	return d.ctx
}

// Cancel cancels the Diameter environment context.
func (d *Diameter) Cancel() {
	d.cancel()
}

// WgAdd increments the wait group counter by delta.
func (d *Diameter) WgAdd(delta int) {
	d.wgDone.Add(delta)
}

// WgDone decrements the wait group counter.
func (d *Diameter) WgDone() {
	d.wgDone.Done()
}

// Wait waits for environment-dependent tasks (such as the server) to complete.
func (d *Diameter) Wait() {
	d.wgDone.Wait()
}

// Logger provides the 'log/slog' instance for environment-dependent tasks (such as the server).
func (d *Diameter) Logger() *slog.Logger {
	return d.logger
}

// Dict returns a reference to the Diameter dictionary.
func (d *Diameter) Dict() *dict.Dict {
	return &d.dict
}

// Peers returns a reference to the peer nodes configuration.
func (d *Diameter) Peers() *node.Nodes {
	return &d.peers
}

// Store returns a reference to the AVP storage.
func (d *Diameter) Store() *AvpStore {
	return &d.store
}

// Mode returns the current Diameter operating mode.
func (d *Diameter) Mode() int32 {
	return d.mode.Load()
}

// SetMode changes the Diameter operating mode.
// Returns an error if the mode is invalid.
func (d *Diameter) SetMode(mode int32) error {
	if mode != ModeTransaction && mode != ModeSession {
		return &diwe.ErrInvalidMode{Mode: mode}
	}
	d.mode.Store(mode)
	return nil
}

// Loaders
//
// LoadDict loads a Diameter dictionary from a file with the specified format.
// It registers codecs for all AVP types defined in the dictionary.
// Returns an error if the file cannot be loaded or parsed.
func (d *Diameter) LoadDict(file string, format int) error {
	if err := d.dict.LoadFromFile(file, format); err != nil {
		return err
	}

	// Register codecs for each AVP type
	d.RegisterCodec(d.dict.AvpDataType().OctetString, mkvOctetString, serOctetString, desOctetString, cpvOctetString, txtOctetString)
	d.RegisterCodec(d.dict.AvpDataType().Integer32, mkvInteger32, serInteger32, desInteger32, cpvInteger32, txtInteger32)
	d.RegisterCodec(d.dict.AvpDataType().Unsigned32, mkvUnsigned32, serUnsigned32, desUnsigned32, cpvUnsigned32, txtUnsigned32)
	d.RegisterCodec(d.dict.AvpDataType().Integer64, mkvInterger64, serInteger64, desInteger64, cpvInteger64, txtInteger64)
	d.RegisterCodec(d.dict.AvpDataType().Unsigned64, mkvUnsigned64, serUnsigned64, desUnsigned64, cpvUnsigned64, txtUnsigned64)
	d.RegisterCodec(d.dict.AvpDataType().Float32, mkvFloat32, serFloat32, desFloat32, cpvFloat32, txtFloat32)
	d.RegisterCodec(d.dict.AvpDataType().Float64, mkvFloat64, serFloat64, desFloat64, cpvFloat64, txtFloat64)
	d.RegisterCodec(d.dict.AvpDataType().Address, mkvIpAddress, serIpAddress, desIpAddress, cpvIpAddress, txtIpAddress)
	d.RegisterCodec(d.dict.AvpDataType().Time, mkvTime, serTime, desTime, cpvTime, txtTime)
	d.RegisterCodec(d.dict.AvpDataType().UTF8String, mkvUTF8String, serUTF8String, desUTF8String, cpvUTF8String, txtUTF8String)
	d.RegisterCodec(d.dict.AvpDataType().Identity, mkvUTF8String, serUTF8String, desUTF8String, cpvUTF8String, txtUTF8String)
	d.RegisterCodec(d.dict.AvpDataType().URI, mkvUTF8String, serUTF8String, desUTF8String, cpvUTF8String, txtUTF8String)
	d.RegisterCodec(d.dict.AvpDataType().IPFilterRule, mkvUTF8String, serUTF8String, desUTF8String, cpvUTF8String, txtUTF8String)
	d.RegisterCodec(d.dict.AvpDataType().QoSFilterRule, mkvUTF8String, serUTF8String, desUTF8String, cpvUTF8String, txtUTF8String)
	d.RegisterCodec(d.dict.AvpDataType().Enumerated, mkvEnumerated, serEnumerated, desEnumerated, cpvEnumerated, txtEnumerated)
	d.RegisterCodec(d.dict.AvpDataType().Grouped, mkvGrouped, serGrouped, desGrouped, cpvGrouped, txtGrouped)

	// Map Diameter types to Go types for value conversion
	d.registerDiaTypes(d.Dict().AvpDataType().Address, reflect.TypeOf(""), reflect.TypeOf(""))
	d.registerDiaTypes(d.Dict().AvpDataType().Enumerated, reflect.TypeOf(""), reflect.TypeOf(0))
	d.registerDiaTypes(d.Dict().AvpDataType().Identity, reflect.TypeOf(""), reflect.TypeOf(""))
	d.registerDiaTypes(d.Dict().AvpDataType().OctetString, reflect.TypeOf(""), reflect.TypeOf(0))
	d.registerDiaTypes(d.Dict().AvpDataType().IPFilterRule, reflect.TypeOf(""), reflect.TypeOf(""))
	d.registerDiaTypes(d.Dict().AvpDataType().QoSFilterRule, reflect.TypeOf(""), reflect.TypeOf(""))
	d.registerDiaTypes(d.Dict().AvpDataType().Time, reflect.TypeOf(""), reflect.TypeOf(""))
	d.registerDiaTypes(d.Dict().AvpDataType().UTF8String, reflect.TypeOf(""), reflect.TypeOf(""))
	d.registerDiaTypes(d.Dict().AvpDataType().URI, reflect.TypeOf(""), reflect.TypeOf(""))
	d.registerDiaTypes(d.Dict().AvpDataType().Integer32, reflect.TypeOf(int32(0)), reflect.TypeOf(0))
	d.registerDiaTypes(d.Dict().AvpDataType().Integer64, reflect.TypeOf(int64(0)), reflect.TypeOf(0))
	d.registerDiaTypes(d.Dict().AvpDataType().Unsigned32, reflect.TypeOf(uint32(0)), reflect.TypeOf(0))
	d.registerDiaTypes(d.Dict().AvpDataType().Unsigned64, reflect.TypeOf(uint64(0)), reflect.TypeOf(0))
	d.registerDiaTypes(d.Dict().AvpDataType().Float32, reflect.TypeOf(float32(0.0)), reflect.TypeOf(0))
	d.registerDiaTypes(d.Dict().AvpDataType().Float64, reflect.TypeOf(float64(0.0)), reflect.TypeOf(0))
	d.registerDiaTypes(d.Dict().AvpDataType().Grouped, reflect.TypeOf([]*Avp{}), nil)

	return nil
}

// LoadPeers loads peer configuration from a file.
// Returns an error if the file cannot be loaded or parsed.
func (d *Diameter) LoadPeers(file string) error {
	return d.peers.LoadFromFile(file, d)
}

// LoadData loads AVP data from a file with the specified append mode.
// Returns an error if the file cannot be loaded or parsed.
func (d *Diameter) LoadData(file string) error {
	return d.store.LoadFromFile(file, AvpStoreAppend, 0)
}

// registerCodec registers codec functions for a specific AVP type.
func (d *Diameter) RegisterCodec(avpType int, mkv MakeValueFn, ser SerializeFn, des DeserializeFn, cpv CopyValueFn, txt ToTextFn) {
	d.codecs[avpType] = CodecFuncs{mkv, ser, des, cpv, txt}
}

// Codec return codec functions for a specific AVP type.
func (d *Diameter) Codec(avpType int) (CodecFuncs, bool) {
	if codec, exists := d.codecs[avpType]; exists {
		return codec, true
	}

	return CodecFuncs{}, false
}

// registerDiaTypes maps Diameter types to Go types.
func (d *Diameter) registerDiaTypes(diaType int, goType1, goType2 reflect.Type) {
	d.dia2go[diaType] = avpGoType{type1: goType1, type2: goType2}
}

// Message handling methods
//
// NewMessage creates a new Diameter message with the specified application ID,
// command ID, and flags. It automatically populates AVPs from the store
// based on command rules.
// Returns an error if the application, command, or required AVPs cannot be resolved.
func (d *Diameter) NewMessage(appId any, cmdId any, request, fetchAvps bool) (*Message, error) {
	app, err := d.dict.GetApp(appId)
	if err != nil {
		return nil, err
	}

	cmd, err := d.dict.GetCmd(cmdId, app)
	if err != nil {
		return nil, err
	}

	m := getMessage()
	m.Version = Version
	m.AppId = app.Id
	m.CmdCode = cmd.Code
	m.Flags = cmd.Flags
	m.Length = MinMessageLen
	m.EndToEnd = d.rng.Uint32()
	m.HopByHop = d.rng.Uint32()
	m.env = d
	m.avps = make([]*Avp, 0)

	if request {
		m.Flags |= d.dict.CmdFlag().R
	}

	if !fetchAvps {
		return m, nil
	}

	var avpRules []dict.AvpRule
	if request {
		avpRules = cmd.Request
	} else {
		avpRules = cmd.Answer
	}

	// Populate AVPs from store based on command rules
	for _, avpRule := range avpRules {
		avpDesc, err := d.dict.GetAvp(avpRule.Name)
		if err != nil {
			return nil, err
		}

		avps := d.store.Fetch(avpDesc.Code)
		if avps == nil {
			if avpRule.Required {
				return nil, &diwe.ErrNoValueForReqAvp{Avp: avpRule.Name}
			}
			continue
		}

		for _, avp := range avps {
			copied, err := avp.Copy()
			if err != nil {
				return nil, err
			}
			if err := m.AddAvp(copied); err != nil {
				return nil, err
			}
			m.Length += alignTo4(avp.Data().Size)
		}
	}

	// Handle Session-Id for transaction mode
	if m.AppId != 0 && d.Mode() == ModeTransaction {
		if err := d.handleSessionId(m); err != nil {
			return nil, err
		}
	}

	return m, nil
}

// handleSessionIdTransaction modifies the Session-Id AVP for transaction mode
// by appending timestamp components.
func (d *Diameter) handleSessionId(m *Message) error {
	sessionId, err := m.GetAvp(avpSessionId)
	if err != nil {
		return nil // No AVP found, it's optional
	}

	values := d.Store().Fetch(avpSessionId)
	if values == nil {
		return &diwe.ErrNoValueForReqAvp{Avp: sessionId}
	}

	now := time.Now().UnixNano()
	hi := now >> 32
	lo := now & mask32bits

	value := values[0].Value()
	if v, ok := value.(string); ok {
		value = fmt.Sprintf("%s;%d;%d", strings.Clone(v), hi, lo)
		if err := sessionId.SetValue(value); err != nil {
			return err
		}
	} else {
		return &diwe.ErrInvalidAvpValue{Avp: sessionId.Name(), Value: value}
	}

	return nil
}

// NewRequest creates a new Diameter request message.
// Returns an error if the message cannot be created.
func (d *Diameter) NewRequest(appId any, cmdCode any) (*Message, error) {
	return d.NewMessage(appId, cmdCode, true, true)
}

// NewAnswer creates a new Diameter answer message.
// Returns an error if the message cannot be created.
func (d *Diameter) NewAnswer(appId any, cmdCode any) (*Message, error) {
	return d.NewMessage(appId, cmdCode, false, true)
}

// NewEmptyMessage creates a new Diameter empty message.
// Returns the message without a error
func (d *Diameter) NewEmptyMessage() *Message {
	m := getMessage()
	m.env = d
	return m
}

// AVP handling methods
//
// NewAvp creates a new AVP with the specified parameters.
// Sets the vendor-specific flag if a vendor ID is provided.
func (d *Diameter) NewAvp(name string, code uint32, flags uint8, vndId uint32, datatype int) *Avp {
	if vndId != 0 {
		flags |= d.dict.AvpFlag().V
	}

	avp := getAvp()
	avp.header = dict.Avp{
		Name:  name,
		Code:  code,
		Flags: flags,
		VndId: vndId,
		Type:  datatype,
	}
	avp.env = d
	avp.value = nil
	return avp
}

// GetAvp retrieves an AVP definition from the dictionary.
// Returns a new empty AVP instance with the dictionary information.
func (d *Diameter) GetAvp(avpId any) (*Avp, error) {
	header, err := d.dict.GetAvp(avpId)
	if err != nil {
		return nil, err
	}

	avp := getAvp()
	avp.header = *header
	avp.env = d
	avp.value = nil
	return avp, nil
}

// The BytesToMessage function converts a byte slice to a Diameter message.
// Returns the Diameter message or an error if the operation fails.
func (d *Diameter) BytesToMessage(data []byte) (*Message, error) {
	if len(data) < int(MinMessageLen) {
		return nil, &diwe.ErrMsgTooShort{Len: len(data)}
	}

	msg := getMessage()
	msg.env = d
	err := msg.Deserialize(data)
	if err != nil {
		putMessage(msg)
		return nil, err
	}

	return msg, nil
}

// Network peers handling methods
//
// NewPeer creates a new peer node with the specified configuration.
// Returns an error if the peer cannot be created.
func (d *Diameter) NewPeer(name string, addr string, port int, proto string, timeout int) (*node.Node, error) {
	return d.peers.NewPeer(name, addr, port, proto, timeout, d)
}

// NewPeerEx creates a new client node with the specified transport connection.
func (d *Diameter) NewPeerEx(tr transport.ITransport, ucb node.UserCallbackFn) (*node.Node, error) {
	return d.peers.NewPeerEx(tr, d, ucb)
}

// SendMessage sends Diameter message to the peer.
// Returns an error if the operation fails.
func (d *Diameter) SendMessage(peer *node.Node, msg *Message) error {
	data, err := msg.Serialize()
	if err != nil {
		return err
	}

	return peer.SendTo(data)
}

// The RecvMessage function receives a Diameter message from the peer.
// Returns the Diameter message or an error if the operation fails.
// wait - wait for a message to be received
func (d *Diameter) RecvMessage(peer *node.Node, wait bool) (*Message, error) {
	var (
		data []byte
		err  error
	)

	data, err = peer.RecvFrom(wait)
	if err != nil {
		return nil, err
	}

	msg, err := d.BytesToMessage(data)
	if err != nil {
		return nil, err
	}

	return msg, nil
}

// The ReplyToMessage function replies to a Diameter message from the peer.
// Returns an error if the operation failed.
func (d *Diameter) ReplyToMessage(peer *node.Node, msg *Message) error {
	reply, err := msg.Response()
	if err != nil {
		return err
	}

	return d.SendMessage(peer, reply)
}

// PCAP handling methods
//
// PcapWriter returns the PCAP writer defined for the Diameter instance.
func (d *Diameter) Pcap() *pcap.Pcap {
	return d.pcap
}

// NewPcapWriter creates the PCAP writer for the Diameter instance.
func (d *Diameter) NewPcapWriter() {
	d.pcap = pcap.New()
}

// PcapOpen opens the PCAP for the Diameter instance.
func (d *Diameter) PcapOpen(file string, append bool) error {
	return d.pcap.Open(file, append)
}

// PcapClose closes the PCAP for the Diameter instance.
func (d *Diameter) PcapClose() error {
	if d.pcap != nil {
		return d.pcap.Close()
	}
	return nil
}

// PcapSync synchronise the PCAP for the Diameter instance.
func (d *Diameter) PcapSync() {
	if d.pcap != nil {
		d.pcap.Sync()
	}
}

// getAvp retrieves an Avp from the pool.
func getAvp() *Avp {
	return avpPool.Get().(*Avp)
}

// putAvp returns an Avp to the pool after resetting it.
func putAvp(avp *Avp) {
	avp.reset()
	avpPool.Put(avp)
}

// getMessage retrieves a Message from the pool.
func getMessage() *Message {
	return messagePool.Get().(*Message)
}

// putMessage returns a Message and all its AVPs to their respective pools.
func putMessage(m *Message) {
	for _, avp := range m.avps {
		putAvp(avp)
	}
	m.reset()
	messagePool.Put(m)
}

// PcapAppend turns on or off appending to the existing PCAP file for the Diameter instance.
func (d *Diameter) PcapAppend(append bool) {
	d.pcap.Append(append)
}

// Debug and info methods
//
// TraceLevel returns the current verbosity level.
func (d *Diameter) TraceLevel() int32 {
	return d.verbLvl.Load()
}

// SetTraceLevel sets the verbosity level for debug output.
func (d *Diameter) SetTraceLevel(level int32) {
	d.verbLvl.Store(level)
}

// Trace prints debug information if the specified level is less than or equal
// to the current verbosity level.
// Object should implement the ITrace interface.
func (d *Diameter) Trace(obj ITrace, level int32) {
	if level > d.TraceLevel() || obj == nil {
		return
	}

	fmt.Println()
	obj.Trace()
	fmt.Println()
}

// IDiameter interface implementation.
//
// The CreateMessage create a Diameter message.
// Returns the byte slice or an error if the operation fails.
func (d *Diameter) CreateMessage(appId uint32, cmdCode uint32, request bool) ([]byte, error) {
	msg, err := d.NewMessage(appId, cmdCode, request, true)
	if err != nil {
		return nil, err
	}

	return msg.Serialize()
}

// The CreateResponse create a Diameter message as response to the given request.
// Returns the byte slice or an error if the operation fails.
func (d *Diameter) CreateResponse(data []byte) ([]byte, error) {
	msg, err := d.BytesToMessage(data)
	if err != nil {
		return nil, err
	}

	reply, err := msg.Response()
	if err != nil {
		return nil, err
	}

	bytes, err := reply.Serialize()
	if err != nil {
		return nil, err
	}

	return bytes, nil
}

// MessageHeader parses a Diameter message header
// Returns the version, length, appId, cmdCode, flags, hopByHop, endToEnd,
// or error if message malformed.
func (d *Diameter) MessageHeader(data []byte) (byte, uint32, uint32, uint32, byte, uint32, uint32, error) {
	if len(data) < int(MinMessageLen) {
		return 0, 0, 0, 0, 0, 0, 0, &diwe.ErrMsgTooShort{Len: len(data)}
	}

	// Byte 0: Version
	version := uint8(data[0])
	// Bytes 0-3: Length (24-bit, mask to get lower 24 bits)
	length := binary.BigEndian.Uint32(data[0:4]) & 0x00FFFFFF
	// Byte 4: Flags
	flags := uint8(data[4])
	// Bytes 4-7: Command Code (24-bit)
	cmdCode := binary.BigEndian.Uint32(data[4:8]) & 0x00FFFFFF
	// Bytes 8-11: Application ID
	appId := binary.BigEndian.Uint32(data[8:12])
	// Bytes 12-15: Hop-by-Hop
	hopByHop := binary.BigEndian.Uint32(data[12:16])
	// Bytes 16-19: End-to-End
	endToEnd := binary.BigEndian.Uint32(data[16:20])

	return version, length, appId, cmdCode, flags, hopByHop, endToEnd, nil
}

// IsCommonMessage returns true if the message is a common Diameter message (App ID == 0).
func (d *Diameter) IsCommonMessage(appId uint32) bool {
	return appId == api.AppIdCommonMessages
}

// IsRequest returns true if the message is a request (flags & 0x80 != 0).
func (d *Diameter) IsRequest(flags byte) bool {
	return flags&d.Dict().CmdFlag().R != 0
}

// GetResultCode returns the result code (AVP 270) from a Diameter message.
func (d *Diameter) GetResultCode(data []byte) (uint32, error) {
	msg, err := d.BytesToMessage(data)
	if err != nil {
		return 0, err
	}

	avpResultCode, err := msg.GetAvp(avpResultCode)
	if err != nil {
		return 0, err
	}

	v, ok := avpResultCode.Value().(uint32)
	if !ok {
		return 0, &diwe.ErrInvalidAvpValue{Avp: avpResultCode, Value: v}
	}

	return v, nil
}

// GetResultCodeEx returns the experimental result code (AVP 297) from a Diameter message.
func (d *Diameter) GetResultCodeEx(data []byte) (uint32, error) {
	return 0, &diwe.ErrNotImplemented{}
}

// TracerMessage traces the message.
func (d *Diameter) TraceMessage(data []byte) {
	msg, err := d.BytesToMessage(data)
	if err != nil {
		return
	}

	d.Trace(msg, TraceCM)
}
