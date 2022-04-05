package server

import (
	"crypto/rand"
	"encoding/hex"
	checkErrors "errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"gitlab.com/NebulousLabs/errors"

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
	wallet        modules.Wallet
	heartbeat     time.Time
}

// StartHTTPServer starts the HTTP server to serve the GUI.
func StartHTTPServer() {
	wg := &sync.WaitGroup{}
	wg.Add(1)
	srv = &http.Server{Addr: ":4300", Handler: buildHTTPRoutes()}
	go func() {
		defer wg.Done()
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			fmt.Printf("Unable to start server: %v\n", err)
			srv = nil
		}
	}()
	waitCh = make(chan struct{})
	go func() {
		wg.Wait()
		close(waitCh)
	}()
}

// IsRunning returns true when the server is running
func IsRunning() bool {
	if srv == nil {
		return false
	}
	for i := 0; i < 20; i++ {
		time.Sleep(5 * time.Millisecond)
		if srv == nil {
			return false
		}
	}
	return srv != nil
}

// Wait returns the servers wait channel
func Wait() chan struct{} {
	return waitCh
}

// AttachNode attaches the node to the HTTP server.
func AttachNode(node *node.Node, params *node.NodeParams) {
	n = node
	nParams = params
	if srv != nil {
		srv.Handler = buildHTTPRoutes()
	}
}

// newWallet creates a new wallet module and attaches it to the node.
func newWallet(walletDirName string, sessionID string) (modules.Wallet, error) {
	walletDir := filepath.Join(n.Dir, "wallets", walletDirName)
	_, err := os.Stat(walletDir)
	if err == nil {
		return nil, fmt.Errorf("%s already exists", walletDirName)
	}
	loadStart := time.Now()
	walletDeps := nParams.WalletDeps
	if walletDeps == nil {
		walletDeps = modules.ProdDependencies
	}
	fmt.Printf("Loading wallet...")
	cs := n.ConsensusSet
	tp := n.TransactionPool
	w, err := wallet.NewCustomWallet(cs, tp, walletDir, walletDeps)
	if err != nil {
		return nil, err
	}
	if w != nil {
		fmt.Println(" done in", time.Since(loadStart).Seconds(), "seconds.")
	}
	for _, session := range sessions {
		if session.id == sessionID {
			if session.wallet != nil {
				session.wallet.Close()
			}
			session.wallet = w
			return w, nil
		}
	}
	return nil, errors.New("session ID was not found")
}

// loadWallet loads the wallet module and attaches it to the node.
func loadWallet(walletDirName string, sessionID string) (modules.Wallet, error) {
	walletDir := filepath.Join(n.Dir, "wallets", walletDirName)
	_, err := os.Stat(walletDir)
	if checkErrors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("%s does not exist", walletDirName)
	}
	return newWallet(walletDirName, sessionID)
}

// closeWallet closes the wallet and detaches it from the node.
func closeWallet(sessionID string) (err error) {
	for _, session := range sessions {
		if session.id == sessionID {
			wallet := session.wallet
			if wallet != nil {
				session.wallet = nil
				fmt.Println("Closing wallet...")
				err = errors.Compose(wallet.Close())
			}
		}
	}
	return err
}

// CloseAllWallets closes all wallets and detaches them from the node.
func CloseAllWallets() (err error) {
	for _, session := range sessions {
		wallet := session.wallet
		if wallet != nil {
			session.wallet = nil
			fmt.Println("Closing wallet...")
			err = errors.Compose(wallet.Close())
		}
	}
	return err
}

func getWallet(sessionID string) (modules.Wallet, error) {
	for _, session := range sessions {
		if session.id == sessionID {
			if session.wallet != nil {
				return session.wallet, nil
			}
		}
	}
	return nil, errors.New("no wallet is attached to the session")
}

// updateHeartbeat updates and returns the heartbeat time.
func updateHeartbeat(sessionID string) time.Time {
	heartbeat = time.Now()
	for _, session := range sessions {
		if session.id == sessionID {
			session.heartbeat = heartbeat
		}
	}
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
