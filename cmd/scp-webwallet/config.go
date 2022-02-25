package main

import (
	"os"
	"path/filepath"
	"runtime"
	"time"

	"gitlab.com/scpcorp/ScPrime/node"
)

// createNodeParams parses the provided config and creates the corresponding
// node params for the server.
func configNodeParams() node.NodeParams {
	params := node.NodeParams{}
	// Set the modules.
	params.CreateGateway = true
	params.CreateConsensusSet = true
	params.CreateTransactionPool = true
	params.CreateWallet = true
	params.CreateDownloader = true
	params.CreateGui = false
	// Parse remaining fields.
	params.Bootstrap = true
	params.SiaMuxTCPAddress = ":4303"
	params.SiaMuxWSAddress = ":4304"
	params.Dir = defaultScPrimeUiDir()
	params.APIaddr = "localhost:4300"
	params.CheckTokenExpirationFrequency = 1 * time.Hour // default
	params.Headless = true
	return params
}

// defaultScPrimeUiDir returns the default data directory of scp-webwallet.
// The values for supported operating systems are:
//
// Linux:   $HOME/.scprime-webwallet
// MacOS:   $HOME/Library/Application Support/ScPrime-WebWallet
// Windows: %LOCALAPPDATA%\ScPrime-WebWallet
func defaultScPrimeUiDir() string {
	switch runtime.GOOS {
	case "windows":
		return filepath.Join(os.Getenv("LOCALAPPDATA"), "ScPrime-WebWallet")
	case "darwin":
		return filepath.Join(os.Getenv("HOME"), "Library", "Application Support", "ScPrime-WebWallet")
	default:
		return filepath.Join(os.Getenv("HOME"), ".scprime-webwallet")
	}
}
