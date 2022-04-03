package main

import (
	"time"

	"gitlab.com/scpcorp/ScPrime/node"
	"gitlab.com/scpcorp/webwallet/build"
)

// createNodeParams parses the provided config and creates the corresponding
// node params for the server.
func configNodeParams() node.NodeParams {
	params := node.NodeParams{}
	// Set the modules.
	params.CreateGateway = true
	params.CreateConsensusSet = true
	params.CreateTransactionPool = true
	// Parse remaining fields.
	params.Bootstrap = true // set to true when the gateway should use the bootstrap peer list
	params.Dir = build.ScPrimeWebWalletDir()
	params.CheckTokenExpirationFrequency = 1 * time.Hour // default
	return params
}
