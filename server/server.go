package server

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"path/filepath"
	"sync"
	"time"

	"gitlab.com/scpcorp/ScPrime/modules"
	"gitlab.com/scpcorp/ScPrime/modules/wallet"
	"gitlab.com/scpcorp/ScPrime/node"
)

var (
	n         *node.Node
	nParams   *node.NodeParams
	srv       *http.Server
	status    string
	heartbeat time.Time
	sessions  []*Session
	waitCh    chan struct{}
)

// Session is a struct that tracks session settings
type Session struct {
	id            string
	alert         string
	collapseMenu  bool
	txHistoryPage int
	cachedPage    string
}

// StartHTTPServer starts the HTTP server to serve the GUI.
func StartHTTPServer() {
	wg := &sync.WaitGroup{}
	wg.Add(1)
	srv = &http.Server{Addr: ":4300", Handler: buildHTTPRoutes()}
	go func() {
		defer wg.Done()
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			fmt.Printf("ListenAndServe(): %v", err)
		}
	}()
	waitCh = make(chan struct{})
	go func() {
		wg.Wait()
		close(waitCh)
	}()
}

// Wait returns the servers wait channel
func Wait() chan struct{} {
	return waitCh
}

// AttachNode attaches the node to the HTTP server.
func AttachNode(node *node.Node, params *node.NodeParams) {
	n = node
	nParams = params
	srv.Handler = buildHTTPRoutes()
}

// attachWallet loads the wallet module and attaches it to the node.
func attachWallet(walletDirName string) error {
	loadStart := time.Now()
	nParams.CreateWallet = true
	walletDeps := nParams.WalletDeps
	if walletDeps == nil {
		walletDeps = modules.ProdDependencies
	}
	fmt.Printf("Loading wallet...")
	cs := n.ConsensusSet
	tp := n.TransactionPool
	dir := n.Dir
	w, err := wallet.NewCustomWallet(cs, tp, filepath.Join(dir, "wallets", walletDirName), walletDeps)
	if err != nil {
		return err
	}
	if w != nil {
		fmt.Println(" done in", time.Since(loadStart).Seconds(), "seconds.")
	}
	n.Wallet = w
	return nil
}

// closeAndDetachWallet closes the wallet and detaches it from the node.
func closeAndDetachWallet() error {
	nParams.CreateWallet = false
	if n == nil || n.Wallet == nil {
		return nil
	}
	err := n.Wallet.Close()
	if err != nil {
		return err
	}
	n.Wallet = nil
	return nil
}

// updateHeartbeat updates and returns the heartbeat time.
func updateHeartbeat() time.Time {
	heartbeat = time.Now()
	return heartbeat
}

// setStatus sets the status.
func setStatus(s string) {
	status = s
}

// addSessionId adds a new session ID to memory.
func addSessionID() string {
	b := make([]byte, 16) //32 characters long
	rand.Read(b)
	session := &Session{}
	session.id = hex.EncodeToString(b)
	session.collapseMenu = true
	session.txHistoryPage = 1
	session.cachedPage = ""
	sessions = append(sessions, session)
	return session.id
}

// sessionIDExists returns true when the supplied session ID exists in memory.
func sessionIDExists(sessionID string) bool {
	for _, session := range sessions {
		if session.id == sessionID {
			return true
		}
	}
	return false
}

// setAlert sets an alert on the session.
func setAlert(alert string, sessionID string) {
	for _, session := range sessions {
		if session.id == sessionID {
			session.alert = alert
		}
	}
}

// hasAlert returns true when the session has an alert.
func hasAlert(sessionID string) bool {
	for _, session := range sessions {
		if session.id == sessionID {
			return session.alert != ""
		}
	}
	return false
}

// popAlert gets the alert from the session and then clears it from the session.
func popAlert(sessionID string) string {
	for _, session := range sessions {
		if session.id == sessionID {
			alert := session.alert
			session.alert = ""
			return alert
		}
	}
	return ""
}

// collapseMenu sets the menu state to collapsed and returns true
func collapseMenu(sessionID string) bool {
	for _, session := range sessions {
		if session.id == sessionID {
			session.collapseMenu = true
		}
	}
	return true
}

// expandMenu sets the menu state to expanded and returns true
func expandMenu(sessionID string) bool {
	for _, session := range sessions {
		if session.id == sessionID {
			session.collapseMenu = false
		}
	}
	return true
}

// menuIsCollapsed returns true when the menu state is collapsed
func menuIsCollapsed(sessionID string) bool {
	for _, session := range sessions {
		if session.id == sessionID {
			return session.collapseMenu
		}
	}
	// default to the menu being expanded just in case
	return false
}

// setTxHistoryPage sets the session's transaction history page and returns true.
func setTxHistoryPage(txHistoryPage int, sessionID string) bool {
	for _, session := range sessions {
		if session.id == sessionID {
			session.txHistoryPage = txHistoryPage
		}
	}
	return true
}

// getTxHistoryPage returns the session's transaction history page or -1 when no session is found.
func getTxHistoryPage(sessionID string) int {
	for _, session := range sessions {
		if session.id == sessionID {
			return session.txHistoryPage
		}
	}
	return -1
}

// cachedPage caches the page without the menu and returns true.
func cachedPage(cachedPage string, sessionID string) bool {
	for _, session := range sessions {
		if session.id == sessionID {
			session.cachedPage = cachedPage
		}
	}
	return true
}

// getCachedPage returns the session's cached page.
func getCachedPage(sessionID string) string {
	for _, session := range sessions {
		if session.id == sessionID {
			return session.cachedPage
		}
	}
	return ""
}
