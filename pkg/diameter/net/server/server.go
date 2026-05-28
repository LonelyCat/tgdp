//
// Project: TGDP - Traffic Generator for Diameter Protocol
// Description: Simple tool for testing and debugging the Diameter protocol
//
// Author: Alexander Kefeli <alexander.kefeli@gmail.com>
//
// File: server.go
// Description: Diameter pkg: simple Diameter server
//

package server

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"sync"
	"sync/atomic"

	"tgdp/pkg/diameter"
	"tgdp/pkg/diameter/diwe"
	"tgdp/pkg/diameter/net/node"
	"tgdp/pkg/diameter/net/transport"

	sctp "github.com/georgeyanev/go-sctp"
)

// Consts
// maxClients is the maximum number of concurrent client connections.
const (
	maxClients = 100
	// maxWorkers is the maximum number of concurrent workers:
	// maxClients + 2 listeners + 1 shutdown handler.
	maxWorkers = maxClients + 2 + 1
)

// Server states represent the lifecycle of the Diameter server.
const (
	StateStopped  = int32(iota) // Server is not running
	StateRunning                // Server is actively accepting connections
	StateShutdown               // Server is in the process of shutting down
)

// Server commands are used to control the server via the control channel.
const (
	Shutdown         = iota // Command to initiate server shutdown
	ShutdownComplete        // Signal that shutdown has completed
)

// Verbosity levels control the verbosity of log output.
// Higher values include all messages from lower levels.
const (
	Quiet = int32(iota) // 0: No output
	Error               // 1: Error messages only
	Warn                // 2: Warning and above
	Info                // 3: Informational messages and above
	Debug               // 4: All messages including debug info
)

// Server is the main struct representing a Diameter server.
// It manages SCTP and TCP listeners, a worker pool for handling connections,
// and maintains server state.
type Server struct {
	wp           *WorkerPool
	sctpListener transport.SctpListener
	tcpListener  transport.TcpListener
	state        atomic.Int32
	verbLevel    atomic.Int32
	autoReply    bool
	mu           sync.Mutex
	ctx          context.Context
	cancel       context.CancelFunc
	env          *diameter.Diameter
}

// WorkerPool manages a pool of worker goroutines for handling concurrent tasks.
// It uses a semaphore pattern to limit the number of concurrent workers.
type WorkerPool struct {
	maxWorkers int
	workers    chan struct{}
	wg         sync.WaitGroup
	mu         sync.Mutex
}

// sctpListener wraps an SCTP network listener.
type sctpListener struct {
	uri      string
	listener *sctp.SCTPListener
}

// tcpListener wraps a TCP network listener.
type tcpListener struct {
	uri      string
	listener *net.TCPListener
}

// Methods
//
// # WorkerPool
//
// Execute attempts to queue a task in the worker pool.
// Returns true if the task was successfully queued, false if the pool is at capacity.
func (wp *WorkerPool) Execute(task func()) bool {
	select {
	case wp.workers <- struct{}{}:
		wp.wg.Add(1)
		go func() {
			task()
			<-wp.workers
			wp.wg.Done()
		}()
		return true
	default:
		return false
	}
}

// Wait blocks until all queued tasks have completed.
func (wp *WorkerPool) Wait() {
	wp.wg.Wait()
}

// Add increments the wait group counter.
func (wp *WorkerPool) Add(delta int) {
	wp.wg.Add(delta)
}

// Done signals that a task has completed, decrementing the wait group counter.
func (wp *WorkerPool) Done() {
	wp.wg.Done()
}

// Server
//
// Context returns the server's context inherited from Diameter environment.
func (s *Server) envContext() context.Context {
	return s.env.Context()
}

// Context returns the server's context for listeners control.
func (s *Server) context() context.Context {
	return s.ctx
}

// Start begins the Diameter server, listening on the specified address for both SCTP and TCP connections.
// The listenAddr parameter specifies the address in "host:port" format.
// The autoReply parameter controls whether the server automatically replies to messages from connected peers.
// Returns an error if the server fails to start or if no listeners can be created.
func (s *Server) Start(listenAddr string, autoReply bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.IsRunning() {
		return nil
	}

	state := StateStopped
	defer func() {
		s.SetState(state)
	}()

	s.ctx, s.cancel = context.WithCancel(context.Background())

	// Initialize the listeners as the IListener interface
	listeners := []transport.IListener{
		&s.sctpListener,
		&s.tcpListener,
	}

	var created int
	for _, l := range listeners {
		if err := l.Create(listenAddr); err != nil {
			s.Verbose(Error, "Listener creation failed", slog.String("type", l.Name()), slog.Any("error", err))
		} else {
			created++
			s.wp.Add(1)
			s.runListener(l)
		}
	}

	if created == 0 {
		s.Verbose(Error, "No listeners created")
		return &diwe.ErrNoListeners{}
	}

	go func() {
		<-s.envContext().Done()
		s.Shutdown()
	}()

	state = StateRunning

	s.env.WgAdd(1)

	return nil
}

// Wait blocks until all workers in the pool have finished their tasks.
// This should be called after Stop to ensure graceful shutdown.
func (s *Server) Wait() {
	s.wp.Wait()
}

// Shutdown performs the actual server shutdown logic.
// It disconnects all peers, closes the listeners, cancels the context, and sets the state to stopped.
func (s *Server) Shutdown() {
	if !s.state.CompareAndSwap(StateRunning, StateStopped) {
		return
	}

	s.Verbose(Info, "Received shutdown signal, gracefully stopping...")

	s.SetAutoreply(false)

	if err := s.sctpListener.Close(); err != nil {
		s.Verbose(Error, "SCTP listener close failed", slog.Any("error", err))
	}

	if err := s.tcpListener.Close(); err != nil {
		s.Verbose(Error, "TCP listener close failed", slog.Any("error", err))
	}

	s.env.Peers().DisconnectAll(true)
	s.cancel()
	s.Wait()

	s.env.WgDone()
}

// State returns the current server state (Stopped, Running, or Shutdown).
func (s *Server) State() int32 {
	return s.state.Load()
}

// SetState atomically sets the server state to the specified value.
func (s *Server) SetState(state int32) {
	s.state.Store(state)
}

// Autoreply sets auto reply mode ON (true) or OFF (false).
func (s *Server) Autoreply() bool {
	return s.autoReply
}

// SetAutoreply sets auto reply mode ON (true) or OFF (false).
func (s *Server) SetAutoreply(onOff bool) {
	s.autoReply = onOff
}

// IsRunning returns true if the server is currently in the Running state.
func (s *Server) IsRunning() bool {
	return s.State() == StateRunning
}

// IsStopped returns true if the server is currently in the Stopped state.
func (s *Server) IsStopped() bool {
	return s.State() == StateStopped
}

// VerboseLevel returns the current verbosity level (Quiet, Error, Warn, Info, or Debug).
func (s *Server) VerboseLevel() int32 {
	return s.verbLevel.Load()
}

// SetVerboseLevel sets the verbosity level for logging output.
// Levels: Quiet (0), Error (1), Warn (2), Info (3), Debug (4).
func (s *Server) SetVerboseLevel(level int32) {
	s.verbLevel.Store(level)
}

// Verbose logs a message at the specified level if the level is greater than or equal to the current verbosity level.
// The msg parameter is the log message, and attrs are optional structured log attributes.
func (s *Server) Verbose(level int32, msg string, attrs ...slog.Attr) {
	if level > s.VerboseLevel() {
		return
	}

	var lvl slog.Level
	switch level {
	case Error:
		lvl = slog.LevelError
	case Warn:
		lvl = slog.LevelWarn
	case Info:
		lvl = slog.LevelInfo
	case Debug:
		lvl = slog.LevelDebug
	default:
		return
	}

	s.env.Logger().LogAttrs(context.Background(), lvl, msg, attrs...)
}

// Dump prints the current server state and listening addresses to stdout.
// This is a convenience method for debugging and status display.
func (s *Server) Dump() {
	if s.mu.TryLock() {
		defer s.mu.Unlock()
	}

	switch s.State() {
	case StateStopped:
		fmt.Println("Server is STOPPED")
	case StateRunning:
		fmt.Println("Server is RUNNING")
		fmt.Print("Auto reply mode is: ")
		if s.autoReply {
			fmt.Println("ON")
		} else {
			fmt.Println("OFF")
		}
		fmt.Println("Listening on:")
		if s.sctpListener.Ready() {
			fmt.Println("  ", s.sctpListener.Uri())
		}
		if s.tcpListener.Ready() {
			fmt.Println("  ", s.tcpListener.Uri())
		}
	default:
		fmt.Println("Server state is UNKNOWN")
	}
}

// runListener starts a goroutine that accepts incoming connections on the given listener.
// The autoReply parameter controls whether connections are handled with automatic message receiving.
// The wg WaitGroup is decremented when the listener is ready or fails to start.
func (s *Server) runListener(listener transport.IListener) {
	go func() {
		defer func() {
			s.Verbose(Info, "Listener stopped", slog.String("listener", listener.Name()))
			s.wp.Done()
		}()

		for {
			select {
			case <-s.envContext().Done():
				s.cancel()

			case <-s.context().Done():
				return

			default:
				tr, err := listener.Accept()
				if err != nil {
					if transport.IsClosedError(err) {
						return
					}
					s.Verbose(Error, "Accept failed", slog.Any("error", err))
				}

				if !s.wp.Execute(func() {
					s.connHandler(tr)
				}) {
					s.Verbose(Warn, "Number of clients exceeded")
					tr.Close() //nolint:errcheck
				}
			}
		}
	}()
}

// connHandler processes an incoming connection from a transport.
// If autoReply is true, server automaticaly replies to messages from the peer.
// If autoReply is false, it is just received the message..
func (s *Server) connHandler(tr transport.ITransport) {
	rAddr := tr.RemoteAddr()

	s.Verbose(Info, "Connected from", slog.String("address", rAddr))

	peer, _ := s.env.NewPeerEx(tr, s.reply)
	if s.VerboseLevel() == Debug {
		s.env.Trace(peer, diameter.TracePeer)
	}

	defer func() {
		s.Verbose(Info, "Disconnected from", slog.String("address", rAddr))
	}()

	for peer.IsOpen() {
		select {
		case <-s.envContext().Done():
			s.cancel()

		case <-s.context().Done():
			return

		case <-peer.Context().Done():
			return
		}
	}
}

// reply sends a reply to the peer if autoReply is enabled.
func (s *Server) reply(data []byte, peer *node.Node) bool {
	if !s.autoReply {
		return false
	}

	msg, err := s.env.BytesToMessage(data)
	if err != nil {
		s.Verbose(Error, "Reply failed", slog.String("peer", peer.Name), slog.Any("error", err))
		return false
	}
	s.env.Trace(msg, diameter.TraceMsg) // FIXME: Remove or comment for better performance

	response, err := msg.Response()
	if err == nil {
		s.env.Trace(response, diameter.TraceMsg) // FIXME: Remove or comment for better performance
		err = s.env.SendMessage(peer, response)
	}
	if err != nil {
		s.Verbose(Error, "Reply failed", slog.String("peer", peer.Name), slog.Any("error", err))
		return false
	}

	return true
}

// Constructors
//
// New creates a new Diameter server with the given Diameter environment and control channel.
// Returns a pointer to the configured Server ready to be started.
func New(env *diameter.Diameter) *Server {
	return &Server{
		wp: &WorkerPool{
			maxWorkers: maxWorkers,
			workers:    make(chan struct{}, maxWorkers),
		},
		env: env,
	}
}
