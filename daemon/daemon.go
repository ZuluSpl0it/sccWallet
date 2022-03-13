package daemon

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"gitlab.com/scpcorp/webwallet/build"
	"gitlab.com/scpcorp/webwallet/server"

	spdBuild "gitlab.com/scpcorp/ScPrime/build"
	"gitlab.com/scpcorp/ScPrime/node"
)

// printVersionAndRevision prints the daemon's version and revision numbers.
func printVersionAndRevision() {
	if build.Version == "" {
		fmt.Println("WARN: compiled ScPrime web wallet without version.")
	} else {
		fmt.Println("ScPrime web wallet v" + build.Version)
	}
	if build.GitRevision == "" {
		fmt.Println("WARN: compiled ScPrime web wallet without version.")
	} else {
		fmt.Println("ScPrime web wallet Git revision " + build.GitRevision)
	}
	if spdBuild.DEBUG {
		fmt.Println("Running ScPrime daemon with debugging enabled")
	}
	if spdBuild.Version == "" {
		fmt.Println("WARN: compiled ScPrime daemon without version.")
	} else {
		fmt.Println("ScPrime daemon v" + spdBuild.Version)
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

// installKillSignalHandler installs a signal handler for os.Interrupt, os.Kill
// and syscall.SIGTERM and returns a channel that is closed when one of them is
// caught.
func installKillSignalHandler() chan os.Signal {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, os.Kill, syscall.SIGTERM)
	return sigChan
}

func startNode(node *node.Node, nodeParams node.NodeParams, loadStart time.Time) {
	err := loadNode(node, nodeParams)
	if err != nil {
		fmt.Println("Server is unable to create the ScPrime node.")
		fmt.Println(err)
		return
	}
	server.AttachNode(node)
	// Print a 'startup complete' message.
	startupTime := time.Since(loadStart)
	fmt.Printf("Finished full startup in %.3f seconds\n", startupTime.Seconds())
	return
}

// StartDaemon uses the config parameters to initialize modules and start the web wallet.
func StartDaemon(nodeParams node.NodeParams) (err error) {
	// Record startup time
	loadStart := time.Now()

	// listen for kill signals
	sigChan := installKillSignalHandler()

	// Print the Version and GitRevision
	printVersionAndRevision()

	// Install a signal handler that will catch exceptions thrown by mmap'd
	// files.
	installMmapSignalHandler()

	// Print a startup message.
	fmt.Println("Loading ScPrime Web Wallet...")

	// Start Server
	server.StartHTTPServer()

	// Start a node
	node := &node.Node{}
	go startNode(node, nodeParams, loadStart)

	select {
	case <-server.Wait():
		fmt.Println("Server was stopped, quitting...")
	case <-sigChan:
		fmt.Println("\rCaught stop signal, quitting...")
	}

	// Close
	if node != nil {
		closeNode(node, nodeParams)
	}
	return nil
}
