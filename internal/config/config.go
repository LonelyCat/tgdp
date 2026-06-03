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
	"strings"
	"tgdp/pkg/diameter"

	"gopkg.in/yaml.v3"
)

// Consts
//

// Default directories names
const (
	userDir     = ".tgdp"
	pklSubdir   = "pkl"
	batchSubdir = "batch"
	yamlSubdir  = "yaml"
)

// Default files names
const (
	confFile     = "config.yaml"
	avpsDataFile = "avps.yaml"
	peersFile    = "peers.yaml"
	pklDictFile  = "dictionary.pkl"
	replAutoRun  = "autorun"
	replHistory  = "history"
)

// Types
//

type Config struct {
	// Diameter protocol parameters
	DiaDictFile string `yaml:"dictionary_file"`
	DiaMode     string `yaml:"diameter_mode"`
	DiaModeId   int32

	// Data files
	AvpsDataFile  string `yaml:"avps_data_file"`
	PeersDataFile string `yaml:"peers_data_file"`

	// Subdirectories
	BatchSubdir string `yaml:"batch_subdir"`
	YamlSubdir  string `yaml:"yaml_subdir"`
}

// Variables
//

var (
	// User data directory - default is ~/.tgdp
	dataDir = filepath.Join(os.Getenv("HOME"), userDir)
	// Config instance
	config Config
)

// Functions
//

func Load(dir string) error {
	SetDataDir(dir)

	var failed bool
	defer func() {
		if failed {
			config.DiaModeId = diameter.ModeTransaction
			config.DiaDictFile = DialDictFile()
			config.AvpsDataFile = AvpsDataFile()
			config.PeersDataFile = PeersDataFile()
			config.BatchSubdir = BatchDir()
			config.YamlSubdir = YamlDir()
		}
	}()

	yamlFile, err := os.ReadFile(ConfigFile())
	if err != nil {
		failed = true
		return err
	}

	err = yaml.Unmarshal([]byte(yamlFile), &config)
	if err != nil {
		return err
	}

	switch strings.ToLower(config.DiaMode) {
	case "transaction":
		config.DiaModeId = diameter.ModeTransaction
	case "session":
		config.DiaModeId = diameter.ModeSession
	default:
		config.DiaModeId = diameter.ModeUnknown
	}

	return nil
}

func SetDataDir(dir string) {
	if len(dir) > 0 {
		dataDir = dir
	}
}

func DataDir() string {
	return dataDir
}

func ConfigFile() string {
	return getConfigPath(confFile)
}

func DiaMode() int32 {
	return config.DiaModeId
}

func DialDictFile() string {
	return getConfigPath(config.DiaDictFile)
}

func AvpsDataFile() string {
	return getConfigPath(config.AvpsDataFile)
}

func PeersDataFile() string {
	return getConfigPath(config.PeersDataFile)
}

func BatchDir() string {
	return getConfigPath(config.BatchSubdir)
}

func YamlDir() string {
	return getConfigPath(config.YamlSubdir)
}

func AutoRunFile() string {
	return getConfigPath(replAutoRun)
}

func HistoryFile() string {
	return getConfigPath(replHistory)
}

func getConfigPath(file string) string {
	return filepath.Join(DataDir(), file)
}
