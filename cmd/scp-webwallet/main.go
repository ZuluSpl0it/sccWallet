package main

import (
	"fmt"
	"os"

	"gitlab.com/scpcorp/webwallet/daemon"
)

// exit codes
// inspired by sysexits.h
const (
	exitCodeGeneral = 1  // Not in sysexits.h, but is standard practice.
	exitCodeUsage   = 64 // EX_USAGE in sysexits.h
)

// die prints its arguments to stderr, then exits the program with the default
// error code.
func die(args ...interface{}) {
	fmt.Fprintln(os.Stderr, args...)
	os.Exit(exitCodeGeneral)
}

// main starts the daemon.
func main() {
	// configure the the node params.
	params := configNodeParams()
	// Launch the GUI
	go launch()
	// Start the ScPrime web wallet daemon.
	// the startDaemon method will only return when it is shutting down.
	err := daemon.StartDaemon(&params)
	if err != nil {
		die(err)
	}
	// Daemon seems to have closed cleanly. Print a 'closed' message.
	fmt.Println("Shutdown complete.")
}
