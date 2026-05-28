//
// Project: TGDP - Traffic Generator for Diameter Protocol
// Description: Simple tool for testing and debugging the Diameter protocol
//
// Author: Alexander Kefeli <alexander.kefeli@gmail.com>
//
// File: pcap.go
// Description: Diameter pkg: writing PCAP files
//

package pcap

import (
	"net"
	"os"
	"time"

	"tgdp/pkg/diameter/diwe"
	"tgdp/pkg/diameter/net/node"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcapgo"
)

// Consts
//
// Direction of the Diameter message.
const (
	DirIncoming = false
	DirOutgoing = true
)

// FileMode constants for PCAP file writing.
const (
	Append   = true
	Truncate = false
)

// Types
//
// Pcap is a struct that holds the PCAP file information and
// provides methods for writing Diameter message data to the file.
type Pcap struct {
	append bool
	file   string
	fd     *os.File
	wr     *pcapgo.Writer
}

// Constructor
//
// New creates a new Pcap instance with the given file path and mode.
func New() *Pcap {
	return &Pcap{
		file:   "",
		append: false,
		fd:     nil,
		wr:     nil,
	}
}

// Methods
//
// Open opens the PCAP file.
func (p *Pcap) Open(file string, append bool) error {
	if file == "" {
		return &diwe.ErrInvalidParam{}
	}

	p.file = file
	p.append = append

	_, err := os.Stat(p.file)
	newFile := os.IsNotExist(err)

	mode := os.O_WRONLY | os.O_CREATE
	if !newFile && p.append {
		mode |= os.O_APPEND
	} else {
		mode |= os.O_TRUNC
	}

	p.fd, err = os.OpenFile(p.file, mode, 0644)
	if err != nil {
		return &diwe.ErrOpenFile{File: p.file, Err: err}
	}

	p.wr = pcapgo.NewWriter(p.fd)
	if !p.append {
		if err := p.wr.WriteFileHeader(65536, layers.LinkTypeEthernet); err != nil {
			p.Close() // nolint: errcheck
			return &diwe.ErrWriteFile{File: p.file, Err: err}
		}
	}

	return nil
}

// Close closes the PCAP file.
func (p *Pcap) Close() error {
	if p.fd != nil {
		err := p.fd.Close()
		p.wr = nil
		p.fd = nil
		return err
	}

	return nil
}

// Sync synchronises the PCAP file.
func (p *Pcap) Sync() {
	if p.fd != nil {
		p.fd.Sync() // nolint: errcheck
	}
}

// Write writes the given Diameter message data to the PCAP file.
// Returns an error if the write operation fails.
func (p *Pcap) Write(data []byte, peer *node.Node, dir bool) error {
	if !p.IsOpen() {
		return nil
	}

	buf := gopacket.NewSerializeBuffer()
	opts := gopacket.SerializeOptions{
		FixLengths:       true,
		ComputeChecksums: false,
	}

	ethLayer := layers.Ethernet{
		EthernetType: layers.EthernetTypeIPv4,
	}
	if dir /* == DirOutgoing */ {
		ethLayer.SrcMAC = peer.RouteInfo.IfaceMac
		ethLayer.DstMAC = net.HardwareAddr{0x00, 0x00, 0x5E, 0x00, 0x00, 0xFF}
	} else {
		ethLayer.SrcMAC = net.HardwareAddr{0x00, 0x00, 0x5E, 0x00, 0x00, 0xFF}
		ethLayer.DstMAC = peer.RouteInfo.IfaceMac
	}

	ipLayer := layers.IPv4{
		Version: 4,
		TTL:     64,
	}

	if dir /* == DirOutgoing */ {
		ipLayer.SrcIP = net.IP(peer.RouteInfo.LocalIp.AsSlice())
		ipLayer.DstIP = net.IP(peer.RouteInfo.RemoteIp.AsSlice())
	} else {
		ipLayer.SrcIP = net.IP(peer.RouteInfo.RemoteIp.AsSlice())
		ipLayer.DstIP = net.IP(peer.RouteInfo.LocalIp.AsSlice())
	}

	var err error
	if peer.Transport().Type() == node.TransportTcp {
		ipLayer.Protocol = layers.IPProtocolTCP

		tcpLayer := layers.TCP{
			Seq:        110,
			Ack:        0,
			DataOffset: 5,
			Window:     14600,
		}

		if dir /* == DirOutgoing */ {
			tcpLayer.SrcPort = layers.TCPPort(peer.LocalPort)
			tcpLayer.DstPort = layers.TCPPort(peer.RemotePort)
		} else {
			tcpLayer.SrcPort = layers.TCPPort(peer.RemotePort)
			tcpLayer.DstPort = layers.TCPPort(peer.LocalPort)
		}

		err = gopacket.SerializeLayers(buf, opts,
			&ethLayer,
			&ipLayer,
			&tcpLayer,
			gopacket.Payload(data),
		)
	} else {
		ipLayer.Protocol = layers.IPProtocolSCTP

		sctpLayer := layers.SCTP{
			VerificationTag: 0,
			Checksum:        0,
		}

		if dir /* == DirOutgoing */ {
			sctpLayer.SrcPort = layers.SCTPPort(peer.LocalPort)
			sctpLayer.DstPort = layers.SCTPPort(peer.RemotePort)
		} else {
			sctpLayer.SrcPort = layers.SCTPPort(peer.RemotePort)
			sctpLayer.DstPort = layers.SCTPPort(peer.LocalPort)
		}

		sctpData := layers.SCTPData{
			Unordered:     false,
			BeginFragment: true,
			EndFragment:   true,
			SCTPChunk: layers.SCTPChunk{
				Type:         layers.SCTPChunkTypeData,
				Length:       uint16(len(data) + 16),
				ActualLength: int(len(data) + 16),
			},
			PayloadProtocol: layers.SCTPPayloadProtocol(46), // 46 - Diameter
		}

		err = gopacket.SerializeLayers(buf, opts,
			&ethLayer,
			&ipLayer,
			&sctpLayer,
			&sctpData,
			gopacket.Payload(data),
		)
	}

	if err != nil {
		return &diwe.ErrSerLayers{Err: err}
	}

	packet := buf.Bytes()

	ci := gopacket.CaptureInfo{
		Timestamp:     time.Now(),
		CaptureLength: len(packet),
		Length:        len(packet),
	}

	if err := p.wr.WritePacket(ci, packet); err != nil {
		return &diwe.ErrWriteFile{File: p.file, Err: err}
	}

	return nil
}

// IsOn returns whether the pcap is currently writing packets.
func (p *Pcap) IsOpen() bool {
	if p == nil {
		return false
	}
	return p.wr != nil
}

// Append turns on appending to the pcap file instead of overwriting it.
func (p *Pcap) Append(append bool) {
	if p != nil {
		p.append = append
	}
}

// IsAppend returns whether the pcap is currently appending to the file.
func (p *Pcap) IsAppend() bool {
	if p == nil {
		return false
	}
	return p.append
}

// SetFile sets the file path for the pcap.
func (p *Pcap) SetFile(file string) {
	if p != nil {
		p.file = file
	}
}

// File returns the file path for the pcap.
func (p *Pcap) File() string {
	if p == nil {
		return ""
	}
	return p.file
}
