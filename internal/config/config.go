//
// Project: TGDP - Traffic Generator for Diameter Protocol
// Description: Simple tool for testing and debugging the Diameter protocol
//
// Author: Alexander Kefeli <alexander.kefeli@gmail.com>
//
// File: config.go
// Description: Global configuration settings
//

package config

import (
	"os"
	"path/filepath"
)

// -- Constants
// --
const (
	dataDir         = ".tgdp"
	confDir         = "configs"
	replBatchDir    = "batch"
	replAutoRunFile = "autorun"
	replHistoryFile = "history"
)

// -- Functions
// --
func DataDir() string {
	return filepath.Join(os.Getenv("HOME"), dataDir)
}

func ConfDir() string {
	return filepath.Join(DataDir(), confDir)
}

func ReplBatchDir() string {
	return filepath.Join(DataDir(), replBatchDir)
}

func ReplAutoRunFile() string {
	return filepath.Join(DataDir(), replAutoRunFile)
}

func ReplHistoryFile() string {
	return filepath.Join(DataDir(), replHistoryFile)
}
