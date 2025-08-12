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

	"tgdp/pkg/diameter/net/node"
	"tgdp/pkg/diameter/net/transport"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcapgo"
)

// -- Functions
// --
func Write(file string, append bool, data []byte, peer *node.Node, request bool) error {
	// FIXME: protect with mutex for MT

	if len(file) == 0 {
		return nil
	}

	if _, err := os.Stat(file); os.IsNotExist(err) {
		append = false
	}

	mode := os.O_WRONLY | os.O_CREATE
	if append {
		mode |= os.O_APPEND
	} else {
		mode |= os.O_TRUNC
	}
	f, err := os.OpenFile(file, mode, 0644)
	if err != nil {
		return &ErrOpenFile{File: file, Err: err}
	}
	defer f.Close() //nolint:errcheck

	w := pcapgo.NewWriter(f)
	if !append {
		if err := w.WriteFileHeader(65536, layers.LinkTypeEthernet); err != nil {
			return &ErrWriteFile{File: file, Err: err}
		}
	}

	buf := gopacket.NewSerializeBuffer()
	opts := gopacket.SerializeOptions{
		FixLengths:       true,
		ComputeChecksums: false,
	}

	ethLayer := layers.Ethernet{
		EthernetType: layers.EthernetTypeIPv4,
	}
	if request {
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

	if request {
		ipLayer.SrcIP = peer.RouteInfo.LocalIp
		ipLayer.DstIP = peer.RouteInfo.RemoteIp
	} else {
		ipLayer.SrcIP = peer.RouteInfo.RemoteIp
		ipLayer.DstIP = peer.RouteInfo.LocalIp
	}

	if _, ok := peer.Tr.(*transport.Tcp); ok {
		ipLayer.Protocol = layers.IPProtocolTCP

		tcpLayer := layers.TCP{
			Seq:        110,
			Ack:        0,
			DataOffset: 5,
			Window:     14600,
		}

		if request {
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

		if request {
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
		return &ErrSerLayers{Err: err}
	}

	packet := buf.Bytes()

	ci := gopacket.CaptureInfo{
		Timestamp:     time.Now(),
		CaptureLength: len(packet),
		Length:        len(packet),
	}

	if err := w.WritePacket(ci, packet); err != nil {
		return &ErrWriteFile{File: file, Err: err}
	}

	return nil
}
