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
	"tgdp/internal/flags"
)

// -- Constants
// --
const (
	dataDir         = ".tgdp"
	confDir         = "configs"
	replBatchDir    = "batch"
	replYamlDir     = "yaml"
	replAutoRunFile = "autorun"
	replHistoryFile = "history"
)

const (
	AvpDataFile = "avp-data.yaml"
	PeersFile   = "peers.yaml"
	DiaDictPkl  = "pkl/dictionary.pkl"
)

// -- Functions
// --
func DataDir() string {
	if len(*flags.C) > 0 {
		return *flags.C

	}
	return filepath.Join(os.Getenv("HOME"), dataDir)
}

func ConfDir() string {
	return filepath.Join(DataDir(), confDir)
}

func BatchDir() string {
	return filepath.Join(DataDir(), replBatchDir)
}

func YamlDir() string {
	return filepath.Join(DataDir(), replYamlDir)
}

func AutoRunFile() string {
	return filepath.Join(DataDir(), replAutoRunFile)
}

func HistoryFile() string {
	return filepath.Join(DataDir(), replHistoryFile)
}
