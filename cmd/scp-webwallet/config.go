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
	params.CreateDownloader = true // set to true when new wallets should bootstrap consensus
	// Parse remaining fields.
	params.Bootstrap = true // set to true when the gateway should use the bootstrap peer list
	params.Dir = defaultScPrimeUIDir()
	params.APIaddr = "localhost:4300"
	params.CheckTokenExpirationFrequency = 1 * time.Hour // default
	return params
}

// defaultScPrimeUiDir returns the default data directory of scp-webwallet.
// The values for supported operating systems are:
//
// Linux:   $HOME/.scprime-webwallet
// MacOS:   $HOME/Library/Application Support/ScPrime-WebWallet
// Windows: %LOCALAPPDATA%\ScPrime-WebWallet
func defaultScPrimeUIDir() string {
	switch runtime.GOOS {
	case "windows":
		return filepath.Join(os.Getenv("LOCALAPPDATA"), "ScPrime-WebWallet")
	case "darwin":
		return filepath.Join(os.Getenv("HOME"), "Library", "Application Support", "ScPrime-WebWallet")
	default:
		return filepath.Join(os.Getenv("HOME"), ".scprime-webwallet")
	}
}
