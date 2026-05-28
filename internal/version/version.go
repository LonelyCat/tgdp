//
// Project: TGDP - Traffic Generator for Diameter Protocol
// Description: Simple tool for testing and debugging the Diameter protocol
//
// Author: Alexander Kefeli <alexander.kefeli@gmail.com>
//
// File: version.go
// Description: Version information
//

package version

import (
	"fmt"
)

// Variables
//

var (
	Version   = "0.0.0"
	BuildDate = "2025-01-01 00:00:00"
)

// Functions
//

func Show() {
	fmt.Println("TGDP: Traffic Generator for Diameter Protocol")
	fmt.Printf("Version: %s\n", Version)
	fmt.Printf("Build date: %s\n", BuildDate)
	fmt.Println()
}
