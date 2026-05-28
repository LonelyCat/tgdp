//
// Project: TGDP - Traffic Generator for Diameter Protocol
// Description: Simple tool for testing and debugging the Diameter protocol
//
// Author: Alexander Kefeli <alexander.kefeli@gmail.com>
//
// File: nodes.go
// Description: Diameter pkg: collection of network nodes/peers
//

package node

import (
	"context"
	"fmt"
	"iter"
	"math/rand/v2"
	"os"
	"slices"
	"strings"
	"sync"

	"tgdp/pkg/diameter/api"
	"tgdp/pkg/diameter/diwe"
	"tgdp/pkg/diameter/net/transport"

	"gopkg.in/yaml.v3"
)

// yamlPeer represents a peer configuration from YAML.
type yamlPeer struct {
	Address   string `yaml:"address"`
	Port      int    `yaml:"port"`
	Transport string `yaml:"transport"`
	Timeout   int    `yaml:"timeout"`
}

// Nodes is a thread-safe collection of peer nodes.
type Nodes struct {
	mu    sync.RWMutex
	nodes []*Node
}

// NewNodes creates a new empty Nodes collection.
func NewNodes() Nodes {
	return Nodes{
		nodes: make([]*Node, 0),
	}
}

// NewPeer creates and adds a new peer to the collection.
// Returns error if peer with same name already exists.
func (n *Nodes) NewPeer(name string, addr string, port int, proto string, timeout int, diaApi api.IDiameter) (*Node, error) {
	if node, _ := n.GetByName(name); node != nil {
		return nil, &diwe.ErrPeerExists{Peer: name}
	}

	if diaApi == nil {
		return nil, &diwe.ErrInvalidParam{}
	}

	tr, err := transport.New(proto)
	if err != nil {
		return nil, err
	}

	node := &Node{}
	node.parent = n
	node.client = false
	node.Name = name
	node.Address = addr
	node.RemotePort = port
	node.LocalPort = rand.IntN(32768) + transport.DefaultPort
	node.Timeout = timeout
	node.tr = tr
	node.diaApi = diaApi

	node.GetRouteInfo() //nolint:errcheck
	node.SetState(StateClosed)

	n.mu.Lock()
	defer n.mu.Unlock()
	n.nodes = append(n.nodes, node)

	return node, nil
}

// NewPeerEx creates a new client node from an accepted transport connection.
// The NewPeerEx is automatically added to the collection.
func (n *Nodes) NewPeerEx(tr transport.ITransport, diaApi api.IDiameter, ucb UserCallbackFn) (*Node, error) {
	if diaApi == nil {
		return nil, &diwe.ErrInvalidParam{}
	}

	node := &Node{}
	node.parent = n
	node.client = true
	node.tr = tr
	node.Address = tr.RemoteAddr()
	node.Name = fmt.Sprintf("peer-%s", node.Address)
	node.diaApi = diaApi
	node.ucb = ucb
	node.ctx = context.Background()

	node.init()

	node.SetState(StateIOpen) // FIXME: WaitCER ?

	n.mu.Lock()
	defer n.mu.Unlock()
	n.nodes = append(n.nodes, node)

	return node, nil
}

// GetByName retrieves a peer by name (case-insensitive).
func (n *Nodes) GetByName(name string) (*Node, error) {
	if n.mu.TryRLock() {
		defer n.mu.RUnlock()
	}

	for _, node := range n.nodes {
		if strings.EqualFold(node.Name, name) {
			return node, nil
		}
	}

	return nil, &diwe.ErrUnknownPeer{Peer: name}
}

// Remove removes a peer from the collection by name.
func (n *Nodes) Remove(name string) {
	if n.mu.TryLock() {
		defer n.mu.Unlock()
	}

	for i, node := range n.nodes {
		if strings.EqualFold(node.Name, name) {
			n.nodes = slices.Delete(n.nodes, i, i+1)
			break
		}
	}
}

// InterruptAll sends an interrupt signal to all nodes in the collection.
func (n *Nodes) InterruptAll() {
	if n.mu.TryRLock() {
		defer n.mu.RUnlock()
	}

	for _, node := range n.nodes {
		node.Interrupt()
	}
}

// DisconnectAll disconnects all nodes in the collection.
// If clients is true, only disconnects client nodes.
func (n *Nodes) DisconnectAll(clients bool) {
	if n.mu.TryLock() {
		defer n.mu.Unlock()
	}

	for i := len(n.nodes) - 1; i >= 0; i-- {
		node := n.nodes[i]
		if clients && !node.IsClient() {
			continue
		}
		node.Disconnect() // nolint: errcheck
	}
}

// Iter returns a sequence of all nodes in the collection.
func (n *Nodes) Iter() iter.Seq[*Node] {
	return func(yield func(*Node) bool) {
		if n.mu.TryRLock() {
			defer n.mu.RUnlock()
		}

		for _, node := range n.nodes {
			if !yield(node) {
				break
			}
		}
	}
}

// Iter2 returns a sequence of (index, node) pairs.
func (n *Nodes) Iter2() iter.Seq2[int, *Node] {
	return func(yield func(int, *Node) bool) {
		if n.mu.TryRLock() {
			defer n.mu.RUnlock()
		}

		for idx, node := range n.nodes {
			if !yield(idx, node) {
				break
			}
		}
	}
}

// LoadFromFile loads peer configurations from a YAML file.
func (n *Nodes) LoadFromFile(yamlFile string, diaApi api.IDiameter) error {
	yamlText, err := os.ReadFile(yamlFile)
	if err != nil {
		return err
	}

	peers := make(map[string]yamlPeer)
	if err = yaml.Unmarshal(yamlText, &peers); err != nil {
		return err
	}

	for name, peer := range peers {
		port := transport.DefaultPort
		if peer.Port != 0 {
			port = peer.Port
		}

		proto := transport.DefaultProtocol
		if peer.Transport != "" {
			proto = peer.Transport
		}

		timeout := transport.DefaultTimeout
		if peer.Timeout != 0 {
			timeout = peer.Timeout
		}

		if _, err := n.NewPeer(name, peer.Address, port, proto, timeout, diaApi); err != nil {
			return err
		}
	}

	return nil
}

// Print prints all nodes to stdout (alias for Dump).
func (nodes *Nodes) Print(shift ...int) {
	nodes.Dump(shift...)
}

// Dump prints all nodes to stdout.
func (nodes *Nodes) Dump(shift ...int) {
	for _, node := range nodes.nodes {
		node.Trace(shift...)
	}
}
