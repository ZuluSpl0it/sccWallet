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
	"gitlab.com/scpcorp/webwallet/modules/browserconfig"
	"gitlab.com/scpcorp/webwallet/modules/consensesbuilder"
	"gitlab.com/scpcorp/webwallet/server"
)

func loadNode(node *node.Node, params *node.NodeParams) error {
	fmt.Println("Loading modules:")
	// Make sure the path is an absolute one.
	dir, err := filepath.Abs(params.Dir)
	if err != nil {
		return err
	}
	node.Dir = dir
	// Configure Browser
	needsShutdown, err := initializeBrowser(params)
	if err != nil {
		return err
	} else if needsShutdown {
		return nil
	}
	// Bootstrap Consensus Set if necessary
	bootstrapConsensusSet(params)
	// Attach Node To Server
	server.AttachNode(node, params)
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
	// Build Consensus Set if necessary
	buildConsensusSet(params)
	// Load Transaction Pool
	err = loadTransactionPool(params, node)
	if err != nil {
		return err
	}
	return nil
}

func closeNode(node *node.Node, params *node.NodeParams) error {
	fmt.Println("Closing modules:")
	params.CreateWallet = false
	params.CreateTransactionPool = false
	consensusbuilder.Close()
	params.CreateConsensusSet = false
	params.CreateGateway = false
	err := node.Close()
	bootstrapper.Close()
	browserconfig.Close()
	return err
}

func initializeBrowser(params *node.NodeParams) (bool, error) {
	loadStart := time.Now()
	fmt.Printf("Initializing browser...")
	time.Sleep(1 * time.Millisecond)
	browserconfig.Start(params.Dir)
	loadTime := time.Since(loadStart).Seconds()
	if browserconfig.Status() == browserconfig.Closed {
		fmt.Println(" closed after", loadTime, "seconds.")
		return true, nil
	}
	if browserconfig.Status() == browserconfig.Failed {
		fmt.Println(" failed after", loadTime, "seconds.")
		return true, nil
	}
	browser, err := browserconfig.Browser(params.Dir)
	if err != nil {
		fmt.Println(" failed after", loadTime, "seconds.")
		return true, err
	}
	if browserconfig.Status() == browserconfig.Initialized {
		fmt.Printf(" browser initialized to %s in %v seconds.\n", browser, loadTime)
		return true, nil
	}
	fmt.Printf(" browser set to %s in %v seconds.\n", browser, loadTime)
	return false, nil
}

func bootstrapConsensusSet(params *node.NodeParams) {
	loadStart := time.Now()
	fmt.Printf("Bootstrapping consensus...")
	time.Sleep(1 * time.Millisecond)
	bootstrapper.Start(params.Dir)
	loadTime := time.Since(loadStart).Seconds()
	if bootstrapper.Progress() == bootstrapper.Skipped {
		fmt.Println(" skipped after", loadTime, "seconds.")
	} else if bootstrapper.Progress() == bootstrapper.Closed {
		fmt.Println(" closed after", loadTime, "seconds.")
	} else {
		fmt.Println(" done in", loadTime, "seconds.")
	}
}

func loadGateway(params *node.NodeParams, node *node.Node) error {
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

func loadConsensusSet(params *node.NodeParams, node *node.Node) error {
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

func buildConsensusSet(params *node.NodeParams) {
	loadStart := time.Now()
	fmt.Printf("Building consensus set...")
	time.Sleep(1 * time.Millisecond)
	consensusbuilder.Start(params.Dir)
	loadTime := time.Since(loadStart).Seconds()
	if consensusbuilder.Progress() == consensusbuilder.Closed {
		fmt.Println(" closed after", loadTime, "seconds.")
	} else {
		fmt.Println(" done in", loadTime, "seconds.")
	}
}

func loadTransactionPool(params *node.NodeParams, node *node.Node) error {
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

// LoadWallet loads the wallet module
func LoadWallet(params *node.NodeParams, node *node.Node, walletDirName string) error {
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
	w, err := wallet.NewCustomWallet(cs, tp, filepath.Join(dir, "wallets", walletDirName), walletDeps)
	if err != nil {
		return err
	}
	if w != nil {
		fmt.Println(" done in", time.Since(loadStart).Seconds(), "seconds.")
	}
	node.Wallet = w
	return nil
}
