//
// Project: TGDP - Traffic Generator for Diameter Protocol
// Description: Simple tool for testing and debugging the Diameter protocol
//
// Author: Alexander Kefeli <alexander.kefeli@gmail.com>
//
// File: flags.go
// Description: Command Line flags
//

package flags

import (
	"flag"
)

var (
	A = flag.Bool("a", false, "Append PCAP file")
	C = flag.String("c", "", "Config files directory")
	D = flag.Bool("d", false, "show Diameter Dictionary")
	G = flag.Bool("g", false, "print debuG info")
	H = flag.Bool("h", false, "show this Help")
	N = flag.Bool("n", false, "do Not send requests to peers")
	S = flag.String("s", "", "run Server")
	V = flag.Int("v", 1, "Verbose output level")
	W = flag.String("w", "", "Write PCAP file")
	Y = flag.Bool("y", false, "verifY Diameter dictionary")
)
