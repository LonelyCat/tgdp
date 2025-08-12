//
// Project: TGDP - Traffic Generator for Diameter Protocol
// Description: Simple tool for testing and debugging the Diameter protocol
//
// Author: Alexander Kefeli <alexander.kefeli@gmail.com>
//
// File: node.go
// Description: Diameter pkg: net node/peer handling
//

package node

import (
	"fmt"
	"iter"
	"log/slog"
	"math/rand/v2"
	"net"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync/atomic"
	"time"

	"tgdp/internal/config"
	"tgdp/pkg/diameter"
	"tgdp/pkg/diameter/net/transport"

	netroute "github.com/libp2p/go-netroute"
	"gopkg.in/yaml.v3"
)

const (
	StateDisconnected = 0
	StateConnected    = 1
	StateReceiving    = 2
	StateError        = 128
)

const (
	KeyName      = "name"
	KeyAddress   = "address"
	KeyPort      = "port"
	KeyTransport = "transport"
	KeyTimeout   = "timeout"
)

// -- Types
// --
type Node struct {
	// Public fields
	Name       string `yaml:"name" required:"true"`
	Address    string `yaml:"address" required:"true"`
	RemotePort int    `yaml:"port" default:"3868"`
	LocalPort  int    `yaml:"local_port" default:"0"`
	Type       string `yaml:"type"`
	Transport  string `yaml:"transport" default:"sctp"`
	Timeout    int    `yaml:"timeout" default:"30"`
	Tr         transport.ITransport
	RouteInfo  RouteInfo
	// Hidden fields
	state  atomic.Uint32
	rxChan chan rxItem
	ccChan chan int
	client bool
}

type RouteInfo struct {
	RemoteIp net.IP
	LocalIp  net.IP
	IfaceId  int
	IfaceMac net.HardwareAddr
	GwIp     net.IP
}

type rxItem struct {
	msg *diameter.Message
	err error
}

// -- Variables
// --
var (
	peers []*Node
)

// -- Functions
// --
func New(name string, addr string, port int, proto string, timeout int) (*Node, error) {
	if node, _ := GetByName(name); node != nil {
		return nil, &ErrPeerExists{Peer: name}
	}

	tr, err := transport.New(proto)
	if err != nil {
		return nil, err
	}

	node := new(Node)
	node.client = false
	node.Name = name
	node.Address = addr
	node.RemotePort = port
	node.LocalPort = rand.IntN(32768) + transport.DEFAULT_PORT
	node.Timeout = timeout
	node.Tr = tr
	node.Transport = tr.Name()
	if err := node.CollectRouteInfo(); err != nil {
		return nil, err
	}
	node.state.Store(StateDisconnected)

	peers = append(peers, node)
	return node, nil
}

func New2(tr transport.ITransport) *Node {
	node := new(Node)
	node.client = true
	node.Tr = tr
	node.Address = tr.RemoteAddr()
	node.Transport = tr.Name()
	node.Name = fmt.Sprintf("peer-%s", node.Address)
	peers = append(peers, node)

	node.state.Store(StateConnected)

	go node.recvHandler()
	time.Sleep(100 * time.Millisecond)

	return node
}

func GetByName(name string) (*Node, error) {
	for _, node := range peers {
		if strings.EqualFold(node.Name, name) {
			return node, nil
		}
	}
	return nil, &ErrUnknownPeer{Peer: name}
}

/* TODO:
func GetByAddr(peers []*Peer, addr string) *Peer {
	return nil
}
*/

func Remove(name string) {
	for i, node := range peers {
		if strings.EqualFold(node.Name, name) {
			peers = slices.Delete(peers, i, i+1)
			break
		}
	}
}

func DisconnectAll() {
	for _, node := range peers {
		_ = node.Disconnect(true, true)
	}
}

func LoadPeers() error {
	path := filepath.Join(config.ConfDir(), config.PeersFile)
	yamlText, err := os.ReadFile(path)
	if err != nil {
		return &ErrReadYaml{File: path, Err: err}
	}

	obj := make(map[string]any)
	if err = yaml.Unmarshal(yamlText, &obj); err != nil {
		return &ErrParseYaml{File: path, Err: err}
	}

	peers = make([]*Node, 0)
	for name := range obj {
		data := obj[name]
		address := data.(map[string]any)[KeyAddress].(string)
		port := transport.DEFAULT_PORT
		if p, exists := data.(map[string]any)[KeyPort]; exists {
			port = p.(int)
		}
		proto := transport.DEFAULT_PROTOCOL
		if p, exists := data.(map[string]any)[KeyTransport]; exists {
			proto = p.(string)
		}
		timeout := transport.DEFAULT_TIMEOUT
		if t, exists := data.(map[string]any)[KeyTimeout]; exists {
			timeout = t.(int)
		}
		_, err := New(name, address, port, proto, timeout)
		if err != nil {
			return (err)
		}
	}

	return nil
}

func Iter() iter.Seq[*Node] {
	return func(yield func(*Node) bool) {
		for _, value := range peers {
			if !yield(value) {
				break
			}
		}
	}
}

func Iter2() iter.Seq2[int, *Node] {
	return func(yield func(int, *Node) bool) {
		for idx, value := range peers {
			if !yield(idx, value) {
				break
			}
		}
	}
}

// -- Methods
// --
func (node *Node) CollectRouteInfo() error {
	nr, err := netroute.New()
	if err != nil {
		return err
	}

	ips, err := net.LookupIP(node.Address)
	if err != nil {
		return err
	}

	var (
		remoteIp net.IP
		found    = false
	)
	for _, ip := range ips {
		if ipv4 := ip.To4(); ipv4 != nil {
			remoteIp = ipv4
			found = true
			break
		}
	}
	if !found {
		return &ErrNoSuitableAddr{Addr: node.Address}
	}

	iface, gwIp, localIp, err := nr.Route(remoteIp)
	if err != nil {
		return err
	}

	node.RouteInfo.RemoteIp = remoteIp
	node.RouteInfo.LocalIp = localIp
	node.RouteInfo.GwIp = gwIp
	node.RouteInfo.IfaceId = iface.Index
	if iface.HardwareAddr != nil {
		node.RouteInfo.IfaceMac = iface.HardwareAddr
	} else {
		node.RouteInfo.IfaceMac = net.HardwareAddr{0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	}

	return nil
}

func (node *Node) Connect(sendCe bool) error {
	if node.IsConnected() {
		return &ErrAlreadyConnected{node.Name}
	}

	ri := &node.RouteInfo
	if err := node.Tr.Connect(ri.RemoteIp, node.RemotePort, ri.LocalIp, node.LocalPort); err != nil {
		return &ErrConnect{Err: err, Peer: node.Name}
	}
	// _ = node.SetTimeout()

	node.state.Store(StateConnected)

	go node.recvHandler()
	time.Sleep(100 * time.Millisecond)

	if sendCe {
		if err := node.SendCapExchange(); err != nil {
			node.Disconnect(false, true) //nolint:errcheck
			return err
		}
	}

	return nil
}

func (node *Node) Disconnect(sendDp, shutdown bool) error {
	defer func() {
		if !shutdown {
			if node.IsClient() {
				Remove(node.Name)
			}
		} else {
			node.state.Store(StateDisconnected)
		}
	}()

	if node.state.And(StateConnected) != StateConnected {
		return &ErrNotConnected{node.Name}
	}

	if sendDp {
		node.SendDisconnectPeer() //nolint:errcheck
	}

	err := node.Tr.Close()
	if err != nil {
		return &ErrDisconnect{Err: err, Peer: node.Name}
	}

	return nil
}

func (node *Node) SetTimeout(timeout ...int) error {
	if timeout != nil {
		node.Timeout = timeout[0]
	}
	return node.Tr.SetTimeout(node.Timeout)
}

func (node *Node) SendTo(msg *diameter.Message) error {
	if buf, err := msg.Serialize(); err != nil {
		return err
	} else {
		return node.SendTo2(buf)
	}
}

func (node *Node) SendTo2(buf []byte) error {
	if !node.IsConnected() {
		return &ErrNotConnected{node.Name}
	}

	if err := node.Tr.Send(buf); err != nil {
		return &ErrSendTo{Err: err, Peer: node.Name}
	}
	return nil
}

func (node *Node) RecvFrom() (*diameter.Message, error) {
	if !node.IsConnected() {
		return nil, &ErrNotConnected{node.Name}
	}

	defer func() {
		node.state.And(^uint32(StateReceiving))
	}()

	node.state.Or(StateReceiving)
	select {
	case <-node.ccChan:
		return nil, &ErrInterrupted{}
	case rxi, ok := <-node.rxChan:
		if !ok {
			return nil, &ErrRecvFrom{Err: node.Tr.Error(), Peer: node.Name}
		}
		if rxi.err != nil {
			return nil, &ErrRecvFrom{Err: rxi.err, Peer: node.Name}
		}
		return rxi.msg, nil
	}
}

func (node *Node) RecvFrom2() (*diameter.Message, error) {
	if !node.IsConnected() {
		return nil, &ErrNotConnected{node.Name}
	}

	buf, err := node.Tr.Recv()
	if err != nil {
		return nil, &ErrRecvFrom{Err: err, Peer: node.Name}
	}
	msg := new(diameter.Message)
	err = msg.Deserialize(buf)
	return msg, err
}

func (node *Node) State() uint32 {
	return node.state.Load()
}

func (node *Node) IsConnected() bool {
	return node.state.Load()&StateConnected == StateConnected
}

func (node *Node) IsReceiving() bool {
	return node.state.Load()&StateReceiving == StateReceiving
}

func (node *Node) IsClient() bool {
	return node.client
}

func (node *Node) SendIntSignal() {
	node.ccChan <- 1
}

func (node *Node) HasData() bool {
	return len(node.rxChan) > 0
}

func (node *Node) SendCapExchange() error {
	return node.SendCommonMsg(257)
}

func (node *Node) SendWatchDog() error {
	return node.SendCommonMsg(280)
}

func (node *Node) SendDisconnectPeer() error {
	return node.SendCommonMsg(282)
}

func (node *Node) SendCommonMsg(cmd uint32) error {
	req, err := diameter.Request(0, cmd)
	if err != nil {
		return err
	}
	diameter.Verbose(req, diameter.VerboseCM)

	err = node.SendTo(req)
	if err != nil {
		return err
	}

	ans, err := node.RecvFrom()
	if err != nil {
		return err
	}
	diameter.Verbose(ans, diameter.VerboseCM)

	avp, err := ans.GetAvp(268 /*"Result-Code"*/)
	if err != nil {
		return err
	}
	rc := avp.GetValue().(uint32)
	if rc != 2001 {
		return &ErrDiameter{rc}
	}

	return nil
}

func (node *Node) ReplyCommonMsg(req *diameter.Message) error {
	diameter.Verbose(req, diameter.VerboseCM)

	ans, err := req.Reply()
	if err != nil {
		return err
	}
	diameter.Verbose(ans, diameter.VerboseCM)

	// TODO: improve error handling
	if err := node.SendTo(ans); err != nil {
		return &ErrSendTo{node.Name, err}
	}

	if req.CmdCode == 282 {
		_ = node.Disconnect(false, true)
	}

	return nil
}

func (node *Node) Dump(shift ...int) {
	fmt.Printf("Peer: %s\n", node.Name)
	fmt.Printf("  Remote Address: %s\n", node.RouteInfo.RemoteIp)
	fmt.Printf("  Local Address: %s\n", node.RouteInfo.LocalIp)
	fmt.Printf("  Remote Port: %d\n", node.RemotePort)
	fmt.Printf("  Local Port: %d\n", node.LocalPort)
	fmt.Printf("  Transport: %s\n", node.Tr.Name())
	fmt.Println()
}

// -- Helpers
// --
func (node *Node) recvHandler() {
	node.rxChan = make(chan rxItem, 10)
	node.ccChan = make(chan int, 1)
	defer func() {
		close(node.rxChan)
		close(node.ccChan)
	}()

	for {
		buf, err := node.Tr.Recv()
		if err != nil {
			node.state.Store(StateDisconnected)
			break
		}
		msg := new(diameter.Message)
		err = msg.Deserialize(buf)

		if err == nil && msg.IsRequest() && msg.AppId == 0 {
			if err := node.ReplyCommonMsg(msg); err != nil {
				slog.Error(err.Error())
			}
			continue
		}

		rxi := rxItem{msg, err}
		node.rxChan <- rxi
	}
}
