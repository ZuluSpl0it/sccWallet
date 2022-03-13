package daemon

import (
	"fmt"
	"path/filepath"
	"time"

	"gitlab.com/scpcorp/ScPrime/modules"
	"gitlab.com/scpcorp/ScPrime/modules/consensus"
	"gitlab.com/scpcorp/ScPrime/modules/gateway"
	"gitlab.com/scpcorp/ScPrime/modules/transactionpool"
	"gitlab.com/scpcorp/ScPrime/modules/wallet"
	"gitlab.com/scpcorp/ScPrime/node"

	"gitlab.com/scpcorp/webwallet/modules/bootstrapper"
	"gitlab.com/scpcorp/webwallet/server"
)

func loadNode(node *node.Node, params node.NodeParams) error {
	fmt.Println("Loading modules:")
	// Make sure the path is an absolute one.
	dir, err := filepath.Abs(params.Dir)
	if err != nil {
		return err
	}
	node.Dir = dir
	// Bootstrap Consensus Set if necessary
	bootstrapConsensusSet(params)
	// Load Gateway.
	err = loadGateway(params, node)
	if err != nil {
		return err
	}
	// Load Consensus Set
	err = loadConsensusSet(params, node)
	if err != nil {
		return err
	}
	// Load Transaction Pool
	err = loadTransactionPool(params, node)
	if err != nil {
		return err
	}
	// Load Wallet
	err = loadWallet(params, node)
	if err != nil {
		return err
	}
	server.AttachNode(node)
	return nil
}

func closeNode(node *node.Node, params node.NodeParams) error {
	fmt.Println("Closing modules:")
	params.CreateWallet = false
	params.CreateTransactionPool = false
	params.CreateConsensusSet = false
	params.CreateGateway = false
	err := node.Close()
	bootstrapper.Close()
	return err
}

func bootstrapConsensusSet(params node.NodeParams) {
	loadStart := time.Now()
	fmt.Printf("Bootstrapping consensus...")
	time.Sleep(1 * time.Millisecond)
	bootstrapper.Start(params.Dir)
	loadTime := time.Since(loadStart).Seconds()
	fmt.Println(" done in", loadTime, "seconds.")
}

func loadGateway(params node.NodeParams, node *node.Node) error {
	loadStart := time.Now()
	if !params.CreateGateway {
		return nil
	}
	if params.RPCAddress == "" {
		params.RPCAddress = "localhost:0"
	}
	gatewayDeps := params.GatewayDeps
	if gatewayDeps == nil {
		gatewayDeps = modules.ProdDependencies
	}
	fmt.Printf("Loading gateway...")
	dir := node.Dir
	g, err := gateway.NewCustomGateway(params.RPCAddress, params.Bootstrap, filepath.Join(dir, modules.GatewayDir), gatewayDeps)
	if err != nil {
		return err
	}
	if g != nil {
		fmt.Println(" done in", time.Since(loadStart).Seconds(), "seconds.")
	}
	node.Gateway = g
	return nil
}

func loadConsensusSet(params node.NodeParams, node *node.Node) error {
	loadStart := time.Now()
	c := make(chan error, 1)
	defer close(c)
	if !params.CreateConsensusSet {
		return nil
	}
	fmt.Printf("Loading consensus set...")
	consensusSetDeps := params.ConsensusSetDeps
	if consensusSetDeps == nil {
		consensusSetDeps = modules.ProdDependencies
	}
	g := node.Gateway
	dir := node.Dir
	cs, errChanCS := consensus.NewCustomConsensusSet(g, params.Bootstrap, filepath.Join(dir, modules.ConsensusDir), consensusSetDeps)
	if err := modules.PeekErr(errChanCS); err != nil {
		return err
	}
	if cs != nil {
		fmt.Println(" done in", time.Since(loadStart).Seconds(), "seconds.")
	}
	node.ConsensusSet = cs
	return nil
}

func loadTransactionPool(params node.NodeParams, node *node.Node) error {
	loadStart := time.Now()
	if !params.CreateTransactionPool {
		return nil
	}
	fmt.Printf("Loading transaction pool...")
	tpoolDeps := params.TPoolDeps
	if tpoolDeps == nil {
		tpoolDeps = modules.ProdDependencies
	}
	cs := node.ConsensusSet
	g := node.Gateway
	dir := node.Dir
	tp, err := transactionpool.NewCustomTPool(cs, g, filepath.Join(dir, modules.TransactionPoolDir), tpoolDeps)
	if err != nil {
		return err
	}
	if tp != nil {
		fmt.Println(" done in", time.Since(loadStart).Seconds(), "seconds.")
	}
	node.TransactionPool = tp
	return nil
}

func loadWallet(params node.NodeParams, node *node.Node) error {
	loadStart := time.Now()
	if !params.CreateWallet {
		return nil
	}
	walletDeps := params.WalletDeps
	if walletDeps == nil {
		walletDeps = modules.ProdDependencies
	}
	fmt.Printf("Loading wallet...")
	cs := node.ConsensusSet
	tp := node.TransactionPool
	dir := node.Dir
	w, err := wallet.NewCustomWallet(cs, tp, filepath.Join(dir, modules.WalletDir), walletDeps)
	if err != nil {
		return err
	}
	if w != nil {
		fmt.Println(" done in", time.Since(loadStart).Seconds(), "seconds.")
	}
	node.Wallet = w
	return nil
}
