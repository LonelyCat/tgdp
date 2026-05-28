//
// Project: TGDP - Traffic Generator for Diameter Protocol
// Description: Simple tool for testing and debugging the Diameter protocol
//
// Author: Alexander Kefeli <alexander.kefeli@gmail.com>
//
// File: node.go
// Description: Diameter pkg: network nodes/peers handling
//

package node

import (
	"context"
	"fmt"
	"math/rand/v2"
	"net"
	"net/netip"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"

	"tgdp/pkg/diameter/api"
	"tgdp/pkg/diameter/diwe"
	"tgdp/pkg/diameter/net/transport"

	netroute "github.com/libp2p/go-netroute"
)

// ConfKey constants are configuration keys for peer settings.
const (
	// ConfKeyName is the configuration key for peer name.
	ConfKeyName = "name"
	// ConfKeyAddress is the configuration key for peer address.
	ConfKeyAddress = "address"
	// ConfKeyPort is the configuration key for peer port.
	ConfKeyPort = "port"
	// ConfKeyTransport is the configuration key for transport protocol.
	ConfKeyTransport = "transport"
	// ConfKeyTimeout is the configuration key for connection timeout.
	ConfKeyTimeout = "timeout"
)

// State constants represent the peer connection state machine states.
const (
	// StateClosed indicates the peer connection is closed.
	StateClosed = int32(iota)
	// StateWaitConnAck indicates waiting for connection acknowledgment.
	StateWaitConnAck
	// StateWaitCEA indicates waiting for Capabilities-Exchange-Answer.
	StateWaitCEA
	// StateIOpen indicates the connection is open as initiator.
	StateIOpen
	// StateROpen indicates the connection is open as responder.
	StateROpen
	// StateSuspect indicates the peer is suspect (unresponsive).
	StateSuspect
	// StateReOpen indicates reconnecting after failure.
	StateReOpen
	// StateShuttingDown indicates the connection is shutting down.
	StateShuttingDown
	// StateWaitData indicates the node waiting for a data arrival.
	StateWaitData
)

// Transport constants represent transport protocol types.
const (
	// TransportUnknown indicates unknown transport protocol.
	TransportUnknown = iota
	// TransportSctp indicates SCTP transport protocol.
	TransportSctp
	// TransportTcp indicates TCP transport protocol.
	TransportTcp
)

// maxMessages is the maximum number of messages in the receive channel.
const maxMessages = 10

// Node represents a Diameter peer connection with network and state information.
type Node struct {
	// Name is the peer identifier.
	Name string
	// Address is the remote IP address or hostname.
	Address string
	// RemotePort is the remote TCP/SCTP port.
	RemotePort int
	// LocalPort is the local TCP/SCTP port.
	LocalPort int
	// Type is the peer type (MME, HSS, ...).
	Type string
	// Timeout is the connection timeout in seconds.
	Timeout int
	// RouteInfo contains network routing details.
	RouteInfo RouteInfo
	// HostName contains Diamter host name.
	HostName string
	// client indicates if this is a client-side connection.
	client bool
	// parent is the Nodes collection this node belongs to.
	parent *Nodes
	// tr is the transport layer interface.
	tr transport.ITransport
	// state is the atomic connection state.
	state atomic.Int32
	// rxChan is the channel for received data.
	rxChan chan rxItem
	// ccChan is the channel for interrupt signals.
	ccChan chan os.Signal
	// mu is the mutex for race conditions avoiding.
	mu sync.Mutex
	// ctx is the context for inform node is shutdown.
	ctx context.Context
	// cancel is the function to cancel the context.
	cancel context.CancelFunc
	// IDiameter is the API calls to process Diameter messages.
	diaApi api.IDiameter
	// User callback function.
	ucb UserCallbackFn
}

// RouteInfo holds network routing information for a peer.
type RouteInfo struct {
	// IfaceId is the network interface index.
	IfaceId int
	// IfaceMac is the network interface MAC address.
	IfaceMac net.HardwareAddr
	// RemoteIp is the remote IP address.
	RemoteIp netip.Addr
	// LocalIp is the local IP address used for connection.
	LocalIp netip.Addr
	// GwIp is the gateway IP address.
	GwIp netip.Addr
}

// rxItem represents a received data item with optional error.
type rxItem struct {
	// data contains the received bytes.
	data []byte
	// err contains any receive error.
	err error
}

// User defined callback function calling on a data receied
type UserCallbackFn func([]byte, *Node) bool

// Methods
//
// # RouteInfo
//
// GwAddr returns the gateway IP address.
func (ri *RouteInfo) GwAddr() netip.Addr {
	return ri.GwIp
}

// LocalAddr returns the local IP address.
func (ri *RouteInfo) LocalAddr() netip.Addr {
	return ri.LocalIp
}

// RemoteAddr returns the remote IP address.
func (ri *RouteInfo) RemoteAddr() netip.Addr {
	return ri.RemoteIp
}

// CollectRouteInfo queries the routing table to determine local interface,
// gateway, and IP addresses for reaching the remote peer.
func (node *Node) GetRouteInfo() error {
	node.Lock()
	defer node.Unlock()

	nr, err := netroute.New()
	if err != nil {
		return &diwe.WarnGetRouteInfoFailed{Peer: node.Name, Err: err}
	}

	ips, err := net.LookupIP(node.Address)
	if err != nil {
		return &diwe.WarnGetRouteInfoFailed{Peer: node.Name, Err: err}
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
		return &diwe.ErrNoSuitableAddr{Peer: node.Name, Addr: node.Address}
	}

	iface, gwIp, localIp, err := nr.Route(remoteIp)
	if err != nil {
		return &diwe.WarnGetRouteInfoFailed{Peer: node.Name, Err: err}
	}

	node.RouteInfo.IfaceId = iface.Index
	if iface.HardwareAddr != nil {
		node.RouteInfo.IfaceMac = iface.HardwareAddr
	} else {
		node.RouteInfo.IfaceMac = net.HardwareAddr{00, 00, 00, 00, 00, 00}
	}

	ip2addr := func(ip net.IP) netip.Addr {
		if addr, ok := netip.AddrFromSlice(ip); ok {
			return addr
		}
		return netip.Addr{}
	}
	node.RouteInfo.RemoteIp = ip2addr(remoteIp)
	node.RouteInfo.LocalIp = ip2addr(localIp)
	node.RouteInfo.GwIp = ip2addr(gwIp)

	return nil
}

// Connect establishes a connection to the peer.
// Sends "Capabilities-Exchange Request" message after connecting.
func (node *Node) Connect() error {
	node.Lock()
	defer node.Unlock()

	if node.IsOpen() {
		return &diwe.ErrAlreadyConnected{Peer: node.Name}
	}

	node.SetState(StateWaitConnAck)

	node.LocalPort = rand.IntN(32768) + transport.DefaultPort
	ri := &node.RouteInfo
	if err := node.tr.Connect(ri.RemoteIp, node.RemotePort, ri.LocalIp, node.LocalPort); err != nil {
		node.SetState(StateClosed)
		return &diwe.ErrConnect{Err: err, Peer: node.Name}
	}

	node.init()

	node.SetState(StateWaitCEA)

	err := node.sendCommonMessage(api.CmdCapabilitiesExchange, true)
	if err != nil {
		node.SetState(StateClosed)
		return err
	}

	node.SetState(StateROpen)

	return nil
}

// init initializes the contextreceive and signal channels.
func (node *Node) init() {
	node.ctx, node.cancel = context.WithCancel(context.Background())

	node.rxChan = make(chan rxItem, maxMessages)
	node.ccChan = make(chan os.Signal, 1)

	ready := make(chan struct{}, 1)
	go node.asyncHandler(ready)
	<-ready
	close(ready)
}

// Disconnect closes the connection to the peer.
// If sendDp is true, sends Disconnect-Peer before closing.
func (node *Node) Disconnect() error {
	node.Lock()
	defer node.Unlock()

	if !node.IsOpen() {
		return &diwe.ErrNotConnected{Peer: node.Name}
	}

	node.SetState(StateShuttingDown)

	node.sendCommonMessage(api.CmdDisconnectPeer, true) // nolint: errcheck

	return node.Close()
}

// Close closes the connection to the peer and cleanup.
// Returns a error encountered during close transport.
func (node *Node) Close() error {
	if node.TryLock() {
		defer node.Unlock()
	}

	if node.IsClosed() {
		return &diwe.ErrNotConnected{Peer: node.Name}
	}

	if node.IsClient() {
		node.parent.Remove(node.Name)
	}

	node.SetState(StateClosed)
	signal.Stop(node.ccChan)
	node.cancel()
	if node.ccChan != nil {
		close(node.ccChan)
		node.ccChan = nil
	}
	if node.rxChan != nil {
		close(node.rxChan)
		node.ccChan = nil
	}

	err := node.tr.Close()
	if err != nil {
		return &diwe.ErrDisconnect{Err: err, Peer: node.Name}
	}

	return nil
}

// SetTimeout applies the configured timeout to the transport layer.
func (node *Node) SetTimeout(timeout int) error {
	node.Timeout = timeout
	return node.tr.SetTimeout(node.Timeout)
}

// SendTo sends raw bytes to the peer.
func (node *Node) SendTo(data []byte) error {
	node.Lock()
	defer node.Unlock()

	return node.sendTo(data)
}

// sendTo sends raw bytes to the peer.
func (node *Node) sendTo(data []byte) error {
	if err := node.tr.Send(data); err != nil {
		return &diwe.ErrSendTo{Err: err, Peer: node.Name}
	}

	return nil
}

// RecvFrom receives data from the peer with lock.
func (node *Node) RecvFrom(wait bool) ([]byte, error) {
	node.Lock()
	defer node.Unlock()

	if !node.IsOpen() {
		return nil, &diwe.ErrNotConnected{Peer: node.Name}
	}

	return node.recvFrom(wait)
}

// recvFrom receives data from the peer with interrupt signal support.
// Returns io.EOF if an error occurs, or InfInterrupted on signal.
func (node *Node) recvFrom(wait bool) ([]byte, error) {
	if !wait && len(node.rxChan) == 0 {
		return nil, &diwe.InfNoDataAvail{Peer: node.Name}
	}

	prevState := node.State()
	node.SetState(StateWaitData)
	defer func() {
		signal.Stop(node.ccChan)
		if !node.IsClosed() {
			node.SetState(prevState)
		}
	}()

	signal.Notify(node.ccChan, os.Interrupt)

	select {
	case <-node.ctx.Done():
		return nil, net.ErrClosed

	case <-node.ccChan:
		return nil, &diwe.InfInterrupted{Peer: node.Name}

	case rxi, ok := <-node.rxChan:
		if !ok {
			rxi.err = &diwe.ErrRecvFrom{Err: node.tr.Error(), Peer: node.Name}
		}

		if rxi.err != nil {
			return nil, rxi.err
		}

		return rxi.data, nil
	}
}

// Context returns the node's context.
func (node *Node) Context() context.Context {
	return node.ctx
}

// Interrupt triggers an interrupt signal to unblock RecvFrom.
func (node *Node) Interrupt() {
	if node.ccChan != nil && node.State() == StateWaitData {
		node.ccChan <- os.Interrupt
	}
}

// Transport returns the transport layer interface.
func (node *Node) Transport() transport.ITransport {
	return node.tr
}

// State returns the current connection state.
func (node *Node) State() int32 {
	return node.state.Load()
}

// SetState sets the connection state.
func (node *Node) SetState(state int32) {
	node.state.Store(state)
}

// IsOpen returns true if the connection is open (StateIOpen or StateROpen).
func (node *Node) IsOpen() bool {
	state := node.State()
	return state == StateIOpen || state == StateROpen || state == StateWaitData
}

// IsShuttingDown returns true if the node/peer is in the process of shutting down.
func (node *Node) IsClosed() bool {
	state := node.State()
	return state == StateClosed
}

// IsClient returns true if this is a client-side connection.
func (node *Node) IsClient() bool {
	return node.client
}

// HasData returns true if there is pending data in the receive channel.
func (node *Node) HasData() bool {
	return len(node.rxChan) > 0
}

// Lock locks the node.
func (node *Node) Lock() {
	node.mu.Lock()
}

// TryLock attempts to lock the node.
func (node *Node) TryLock() bool {
	return node.mu.TryLock()
}

// Unlock unlocks the node.
func (node *Node) Unlock() {
	node.mu.Unlock()
}

// Trace prints peer information to stdout.
func (node *Node) Trace(shift ...int) {
	fmt.Printf("Peer: %s\n", node.Name)
	fmt.Printf("  Remote Address: %s\n", node.RouteInfo.RemoteAddr())
	fmt.Printf("  Local Address: %s\n", node.RouteInfo.LocalAddr())
	fmt.Printf("  Gateway Address: %s\n", node.RouteInfo.GwAddr())
	fmt.Printf("  Remote Port: %d\n", node.tr.RemotePort())
	fmt.Printf("  Local Port: %d\n", node.tr.LocalPort())
	fmt.Printf("  Transport: %s\n", node.tr.Name())
	fmt.Printf("  State: %d\n", node.State())
	fmt.Println()
}

// recvHandler handles incoming data from the transport layer.
func (node *Node) asyncHandler(ready chan struct{}) {
	ready <- struct{}{}

	for {
		select {
		case <-node.ctx.Done():
			return

		default:
			data, err := node.tr.Recv()
			if err != nil {
				if transport.IsClosedError(err) {
					node.Close() // nolint: errcheck
					return

				}
				node.rxChan <- rxItem{nil, err}
				continue
			}

			if node.handleCommonMessage(data) {
				continue
			}

			if node.ucb != nil && node.ucb(data, node) {
				continue
			}

			node.rxChan <- rxItem{data, err}
		}
	}
}

// handleCommonMessage auto handles incoming common messages (AppID == 0).
func (node *Node) handleCommonMessage(data []byte) bool {
	_, _, appId, _, flags, _, _, err := node.diaApi.MessageHeader(data)
	if err != nil {
		return false
	}

	if node.diaApi.IsCommonMessage(appId) {
		node.diaApi.TraceMessage(data) // FIXME: Remove or comment for better performance

		if node.diaApi.IsRequest(flags) {
			return node.replyCommonMessage(data) == nil
		}
		node.rxChan <- rxItem{data, nil}
		return true
	}

	return false
}

// sendCommonMessage sends a common message (AppID == 0) with the given command code.
// Returns an error if one occurs.
func (node *Node) replyCommonMessage(data []byte) error {
	response, err := node.diaApi.CreateResponse(data)
	if err != nil {
		return err
	}
	node.diaApi.TraceMessage(response) // FIXME: Remove or comment for better performance

	return node.sendTo(response)
}

// sendCommonMessage sends a common message (AppID == 0) with the given command code.
// Returns an error if one occurs.
func (node *Node) sendCommonMessage(cmdCode uint32, request bool) error {
	bytes, err := node.diaApi.CreateMessage(api.AppIdCommonMessages, cmdCode, request)
	if err != nil {
		return err
	}
	node.diaApi.TraceMessage(bytes) // FIXME:  Remove or comment for better performance

	err = node.sendTo(bytes)
	if err != nil {
		return err
	}

	bytes, err = node.recvFrom(true)
	if err != nil {
		return err
	}

	rc, err := node.diaApi.GetResultCode(bytes)
	if err != nil {
		return err
	}

	if rc != api.DiameterSuccess {
		return &diwe.ErrDiameter{Code: rc}
	}

	return nil
}
