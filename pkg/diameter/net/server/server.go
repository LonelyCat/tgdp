//
// Project: TGDP - Traffic Generator for Diameter Protocol
// Description: Simple tool for testing and debugging the Diameter protocol
//
// Author: Alexander Kefeli <alexander.kefeli@gmail.com>
//
// File: server.go
// Description: Diameter pkg: net server handling
//

package server

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"sync"
	"time"

	"tgdp/pkg/diameter"
	"tgdp/pkg/diameter/net/node"
	"tgdp/pkg/diameter/net/transport"

	"github.com/ishidawataru/sctp"
)

// -- Constants
// --
const (
	maxWorkers = 100
)

const (
	StateStopped = 0
	StateRunning = 1
)

// -- Types
// --
type State struct {
	State    int
	SctpInfo string
	TcpInfo  string
}

type WorkerPool struct {
	maxWorkers int
	workers    chan struct{}
	wg         sync.WaitGroup
}

// -- Variables
// --
var (
	state    State
	stopChan chan int = make(chan int, 1)
)

// -- Functions
// --
func NewWorkerPool(maxWorkers int) *WorkerPool {
	return &WorkerPool{
		maxWorkers: maxWorkers,
		workers:    make(chan struct{}, maxWorkers),
	}
}

func (wp *WorkerPool) Execute(task func()) {
	wp.workers <- struct{}{}
	wp.wg.Add(1)
	go func() {
		defer wp.wg.Done()              //nolint:errcheck
		defer func() { <-wp.workers }() //nolint:errcheck
		task()
	}()
}

func (wp *WorkerPool) Wait() {
	wp.wg.Wait()
}

func handleConnection[T transport.ITransport](tr T, cliMode bool, wg *sync.WaitGroup) {
	peer := node.New2(tr)
	diameter.Verbose(peer, diameter.VerbosePeer)
	if !cliMode {
		return
	}
	defer func() {
		peer.Disconnect(true, false) //nolint:errcheck
		wg.Done()
	}()

	for {
		msg, err := peer.RecvFrom()
		if err != nil {
			if _, ok := err.(*node.ErrRecvFrom); !ok {
				slog.Error(err.Error())
			}
			break
		}
		diameter.Verbose(msg, diameter.VerboseMsg)

		msg, err = msg.Reply()
		if err != nil {
			slog.Error(err.Error())
			break
		}
		err = peer.SendTo(msg)
		if err != nil {
			slog.Error(err.Error())
			break
		}
		diameter.Verbose(msg, diameter.VerboseMsg)
	}
}

func Start(listenAddr string, cliMode bool, sema *chan struct{}) {
	riseSema := func() {
		if sema != nil {
			*sema <- struct{}{}
		}
	}

	defer riseSema()

	maxWorkers := maxWorkers

	workerPool := NewWorkerPool(maxWorkers)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() //nolint:errcheck

	var (
		tcpListener  *net.TCPListener
		sctpListener *sctp.SCTPListener
	)

	if addr, err := net.ResolveTCPAddr("tcp", listenAddr); err != nil {
		slog.Error(err.Error())
	} else if tcpListener, err = net.ListenTCP("tcp", addr); err != nil {
		slog.Error(err.Error())
	}

	if addr, err := sctp.ResolveSCTPAddr("sctp", listenAddr); err != nil {
		slog.Error(err.Error())
	} else if sctpListener, err = sctp.ListenSCTP("sctp", addr); err != nil {
		slog.Error(err.Error())
	}

	if sctpListener == nil && tcpListener == nil {
		slog.Error("No listeners created")
		return
	}

	shutdown := false
	go func() {
		<-stopChan
		fmt.Println("\nReceived shutdown signal, gracefully stopping...")
		shutdown = true
		sctpListener.Close() //nolint:errcheck
		tcpListener.Close()  //nolint:errcheck
		cancel()
	}()

	fmt.Println()

	// TCP listener
	workerPool.wg.Add(1)
	go func() {
		defer func() {
			workerPool.wg.Done() //nolint:errcheck
			fmt.Println("TCP listener stopped")
		}()

		if tcpListener == nil {
			return
		}

		state.TcpInfo = fmt.Sprintf("tcp://%v", tcpListener.Addr())

		for {
			select {
			case <-ctx.Done():
				return
			default:
				if conn, err := tcpListener.AcceptTCP(); err != nil {
					if shutdown {
						return
					}
					slog.Error(err.Error())
				} else {
					tcpTr := &transport.Tcp{Connection: conn, Err: nil}
					workerPool.Execute(func() {
						workerPool.wg.Add(1)
						handleConnection(tcpTr, cliMode, &workerPool.wg)
					})
				}
			}
		}
	}()

	// SCTP listener
	workerPool.wg.Add(1)
	go func() {
		defer func() {
			workerPool.wg.Done() //nolint:errcheck
			fmt.Println("SCTP listener stopped")
		}()

		if sctpListener == nil {
			return
		}

		state.SctpInfo = fmt.Sprintf("sctp://%v", sctpListener.Addr())

		for {
			select {
			case <-ctx.Done():
				return
			default:
				if conn, err := sctpListener.AcceptSCTP(); err != nil {
					if shutdown {
						return
					}
					slog.Error(err.Error())
				} else {
					sctpTr := &transport.Sctp{Connection: conn, Err: nil}
					workerPool.Execute(func() {
						workerPool.wg.Add(1)
						handleConnection(sctpTr, cliMode, &workerPool.wg)
					})
				}
			}
		}
	}()

	time.Sleep(100 * time.Millisecond)
	state.State = StateRunning
	riseSema()
	if cliMode {
		ShowState()
	}
	workerPool.Wait()
	state.State = StateStopped
}

func Stop() {
	if state.State == StateRunning {
		stopChan <- 1
		time.Sleep(200 * time.Millisecond)
	}
	state.State = StateStopped
}

func Status() *State {
	return &state
}

func ShowState() {
	switch state.State {
	case StateStopped:
		fmt.Println("Server is stopped")
	case StateRunning:
		fmt.Println("Server is running")
		fmt.Println("Listening on:")
		fmt.Println("  ", state.SctpInfo)
		fmt.Println("  ", state.TcpInfo)
	default:
		fmt.Println("Server state is unknown")
	}
}
