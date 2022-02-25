package main

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"gitlab.com/scpcorp/webwallet/server"

	"gitlab.com/scpcorp/ScPrime/build"
	"gitlab.com/scpcorp/ScPrime/modules"
	"gitlab.com/scpcorp/ScPrime/node"
)

var n *node.Node

// printVersionAndRevision prints the daemon's version and revision numbers.
func printVersionAndRevision() {
	if build.DEBUG {
		fmt.Println("Running with debugging enabled")
	}
	if build.Version == "" {
		fmt.Println("WARN: compiled without version.")
	} else {
		fmt.Println("ScPrime Web Wallet v" + build.Version)
	}
	if build.GitRevision == "" {
		fmt.Println("WARN: compiled without build commit.")
	} else {
		fmt.Println("Git Revision " + build.GitRevision)
	}
}

// installMmapSignalHandler installs a signal handler for Mmap related signals
// and exits when such a signal is received.
func installMmapSignalHandler() {
	// NOTE: ideally we would catch SIGSEGV here too, since that signal can
	// also be thrown by an mmap I/O error. However, SIGSEGV can occur under
	// other circumstances as well, and in those cases, we will want a full
	// stack trace.
	mmapChan := make(chan os.Signal, 1)
	signal.Notify(mmapChan, syscall.SIGBUS)
	go func() {
		<-mmapChan
		fmt.Println("A fatal I/O exception (SIGBUS) has occurred.")
		fmt.Println("Please check your disk for errors.")
		os.Exit(1)
	}()
}

func startNode(nodeParams node.NodeParams, loadStart time.Time) {
	node, errChan := node.New(nodeParams, loadStart)
	fmt.Println("ACTUALLY, THE API IS NOT LOADED. THE LOG ABOVE MESSAGE IS IN THE WRONG GOLANG FILE.")
	if err := modules.PeekErr(errChan); err != nil {
		fmt.Println("server is unable to create the ScPrime node")
	}
	server.AttachNode(node)
	// Print a 'startup complete' message.
	startupTime := time.Since(loadStart)
	fmt.Printf("Finished full startup in %.3f seconds\n", startupTime.Seconds())
	// attach node to daemon
	n = node
}

// installKillSignalHandler installs a signal handler for os.Interrupt, os.Kill
// and syscall.SIGTERM and returns a channel that is closed when one of them is
// caught.
func installKillSignalHandler() chan os.Signal {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, os.Kill, syscall.SIGTERM)
	return sigChan
}

// startDaemon uses the config parameters to initialize modules and start the web wallet.
func startDaemon() (err error) {
	// Record startup time
	loadStart := time.Now()

	// Print the Version and GitRevision
	printVersionAndRevision()

	// Install a signal handler that will catch exceptions thrown by mmap'd
	// files.
	installMmapSignalHandler()

	// Print a startup message.
	fmt.Println("Loading ScPrime Web Wallet...")

	// configure the the node params.
	nodeParams := configNodeParams()

	// Start a node
	go startNode(nodeParams, loadStart)

	// Launch the GUI
	go launch("http://" + nodeParams.APIaddr)

	// Start Server
	httpServerExitDone := &sync.WaitGroup{}
	httpServerExitDone.Add(1)
	server.StartHttpServer(nodeParams.APIaddr, httpServerExitDone)
	httpServerExitDone.Wait()

	// Close
	if n != nil {
		n.Close()
	}
	return nil
}
