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

// Constants
//

const (
	dataDir         = ".tgdp"
	replBatchDir    = "batch"
	replYamlDir     = "yaml"
	replAutoRunFile = "autorun"
	replHistoryFile = "history"
)

const (
	confFile    = "config.yaml"
	avpDataFile = "data/avp-data.yaml"
	peersFile   = "data/peers.yaml"
	diaPklFile  = "pkl/dictionary.pkl"
)

const (
	SessionIdAuto   = 0
	SessionIdManual = 1
)

var (
	SessionIdMode = SessionIdAuto
)

// Functions
//

func DataDir() string {
	if len(*flags.C) > 0 {
		return *flags.C

	}
	return filepath.Join(os.Getenv("HOME"), dataDir)
}

func BatchDir() string {
	return filepath.Join(DataDir(), replBatchDir)
}

func YamlDir() string {
	return filepath.Join(DataDir(), replYamlDir)
}

func getConfigPath(fileName string) string {
	return filepath.Join(DataDir(), fileName)
}

func ConfigFile() string {
	return getConfigPath(confFile)
}

func DiaDictFile() string {
	return getConfigPath(diaPklFile)
}

func AvpDataFile() string {
	return getConfigPath(avpDataFile)
}

func PeersFile() string {
	return getConfigPath(peersFile)
}

func AutoRunFile() string {
	return getConfigPath(replAutoRunFile)
}

func HistoryFile() string {
	return getConfigPath(replHistoryFile)
}
