package daemon

import (
	"fmt"
	"path/filepath"
	"time"

	"gitlab.com/scpcorp/ScPrime/modules"
	"gitlab.com/scpcorp/ScPrime/modules/consensus"
	"gitlab.com/scpcorp/ScPrime/modules/downloader"
	"gitlab.com/scpcorp/ScPrime/modules/gateway"
	"gitlab.com/scpcorp/ScPrime/modules/transactionpool"
	"gitlab.com/scpcorp/ScPrime/modules/wallet"
	"gitlab.com/scpcorp/ScPrime/node"

	"gitlab.com/scpcorp/webwallet/server"
)

func newNode(params node.NodeParams) (*node.Node, error) {
	numModules := params.NumModules()
	i := 0
	fmt.Println("Starting modules:")
	// Make sure the path is an absolute one.
	dir, err := filepath.Abs(params.Dir)
	if err != nil {
		return nil, err
	}
	// Downloader
	loadStart := time.Now()
	d := func() modules.Downloader {
		if !params.CreateDownloader {
			return nil
		}
		i++
		fmt.Printf("(%d/%d) Downloading consensus...", i, numModules)
		return downloader.New(params.Dir)
	}()
	if d != nil {
		loadTime := time.Since(loadStart).Seconds()
		if loadTime < .0001 {
			loadTime = .0001
		}
		fmt.Println(" done in", loadTime, "seconds.")
	}
	// Gateway.
	loadStart = time.Now()
	g, err := func() (modules.Gateway, error) {
		if !params.CreateGateway {
			return nil, nil
		}
		if params.RPCAddress == "" {
			params.RPCAddress = "localhost:0"
		}
		gatewayDeps := params.GatewayDeps
		if gatewayDeps == nil {
			gatewayDeps = modules.ProdDependencies
		}
		i++
		fmt.Printf("(%d/%d) Loading gateway...", i, numModules)
		return gateway.NewCustomGateway(params.RPCAddress, params.Bootstrap, filepath.Join(dir, modules.GatewayDir), gatewayDeps)
	}()
	if err != nil {
		return nil, err
	}
	if g != nil {
		fmt.Println(" done in", time.Since(loadStart).Seconds(), "seconds.")
	}
	// Consensus
	loadStart = time.Now()
	cs, errChanCS := func() (modules.ConsensusSet, <-chan error) {
		c := make(chan error, 1)
		defer close(c)
		if !params.CreateConsensusSet {
			return nil, c
		}
		i++
		fmt.Printf("(%d/%d) Loading consensus...", i, numModules)
		consensusSetDeps := params.ConsensusSetDeps
		if consensusSetDeps == nil {
			consensusSetDeps = modules.ProdDependencies
		}
		return consensus.NewCustomConsensusSet(g, params.Bootstrap, filepath.Join(dir, modules.ConsensusDir), consensusSetDeps)
	}()
	if err := modules.PeekErr(errChanCS); err != nil {
		return nil, err
	}
	if cs != nil {
		fmt.Println(" done in", time.Since(loadStart).Seconds(), "seconds.")
	}
	// Transaction Pool
	loadStart = time.Now()
	tp, err := func() (modules.TransactionPool, error) {
		if !params.CreateTransactionPool {
			return nil, nil
		}
		i++
		fmt.Printf("(%d/%d) Loading transaction pool...", i, numModules)
		tpoolDeps := params.TPoolDeps
		if tpoolDeps == nil {
			tpoolDeps = modules.ProdDependencies
		}
		return transactionpool.NewCustomTPool(cs, g, filepath.Join(dir, modules.TransactionPoolDir), tpoolDeps)
	}()
	if err != nil {
		return nil, err
	}
	if tp != nil {
		fmt.Println(" done in", time.Since(loadStart).Seconds(), "seconds.")
	}
	// Wallet
	loadStart = time.Now()
	w, err := func() (modules.Wallet, error) {
		if !params.CreateWallet {
			return nil, nil
		}
		walletDeps := params.WalletDeps
		if walletDeps == nil {
			walletDeps = modules.ProdDependencies
		}
		i++
		fmt.Printf("(%d/%d) Loading wallet...", i, numModules)
		return wallet.NewCustomWallet(cs, tp, filepath.Join(dir, modules.WalletDir), walletDeps)
	}()
	if err != nil {
		return nil, err
	}
	if w != nil {
		fmt.Println(" done in", time.Since(loadStart).Seconds(), "seconds.")
	}
	node := &node.Node{
		ConsensusSet:    cs,
		Gateway:         g,
		TransactionPool: tp,
		Wallet:          w,
		Downloader:      d,
		Dir:             dir,
	}
	server.AttachNode(node)
	return node, nil
}
