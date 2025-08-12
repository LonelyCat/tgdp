//
// Project: TGDP - Traffic Generator for Diameter Protocol
// Description: Simple tool for testing and debugging the Diameter protocol
//
// Author: Alexander Kefeli <alexander.kefeli@gmail.com>
//
// File: iface.go
// Description: Show interfaces information
//

package iface

import (
	"fmt"
	"log/slog"
	"net"
	"os"

	"github.com/google/gopacket/pcap"
)

func List() {
	ifaces, err := pcap.FindAllDevs()

	if err != nil {
		slog.Error(err.Error())
		return
	}

	fmt.Println("Interfaces:")
	for _, iface := range ifaces {
		fmt.Printf("%s", iface.Name)
		ip := ifaceIp(iface.Name)
		if ip != nil {
			fmt.Printf(": %s", ip)
		}
		fmt.Println()
	}

	os.Exit(1)
}

func ifaceIp(iface string) net.IP {
	i, err := net.InterfaceByName(iface)
	if err != nil {
		return nil
	}
	addrs, err := i.Addrs()
	if err != nil {
		return nil
	}

	var ipAddr net.IP
	for _, addr := range addrs {
		if ipAddr = addr.(*net.IPNet).IP.To4(); ipAddr != nil {
			break
		}
	}

	return ipAddr
}
