package server

import (
  "crypto/rand"
  "encoding/hex"
  "fmt"
	"net/http"
  "sync"
  "time"

  "gitlab.com/scpcorp/ScPrime/node"
)

var (
  n *node.Node
  srv *http.Server
  status string
  heartbeat time.Time
  sessions []*Session
)

// Session is a struct that tracks session settings
type Session struct {
	id            string
	collapseMenu  bool
	txHistoryPage int
	cachedPage    string
}

func StartHttpServer(addr string, wg *sync.WaitGroup) *http.Server {
  srv = &http.Server{Addr: addr, Handler: buildHttpRoutes()}
  go func() {
    defer wg.Done()
    if err := srv.ListenAndServe(); err != http.ErrServerClosed {
      fmt.Printf("ListenAndServe(): %v", err)
    }
  }()
  return srv
}

func AttachNode(node *node.Node) {
  n = node
  srv.Handler = buildHttpRoutes()
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
func addSessionId() string {
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

// sessionIdExists returns true when the supplied session ID exists in memory.
func sessionIdExists(sessionId string) bool {
	for _, session := range sessions {
		if session.id == sessionId {
			return true
		}
	}
	return false
}

// collapseMenu sets the menu state to collapsed and returns true
func collapseMenu(sessionId string) bool {
	for _, session := range sessions {
		if session.id == sessionId {
			session.collapseMenu = true
		}
	}
	return true
}

// expandMenu sets the menu state to expanded and returns true
func expandMenu(sessionId string) bool {
	for _, session := range sessions {
		if session.id == sessionId {
			session.collapseMenu = false
		}
	}
	return true
}

// menuIsCollapsed returns true when the menu state is collapsed
func menuIsCollapsed(sessionId string) bool {
	for _, session := range sessions {
		if session.id == sessionId {
			return session.collapseMenu
		}
	}
	// default to the menu being expanded just in case
	return false
}

// setTxHistoryPage sets the session's transaction history page and returns true.
func setTxHistoryPage(txHistoryPage int, sessionId string) bool {
	for _, session := range sessions {
		if session.id == sessionId {
			session.txHistoryPage = txHistoryPage
		}
	}
	return true
}

// getTxHistoryPage returns the session's transaction history page or -1 when no session is found.
func getTxHistoryPage(sessionId string) int {
	for _, session := range sessions {
		if session.id == sessionId {
			return session.txHistoryPage
		}
	}
	return -1
}

// cachedPage caches the page without the menu and returns true.
func cachedPage(cachedPage string, sessionId string) bool {
	for _, session := range sessions {
		if session.id == sessionId {
			session.cachedPage = cachedPage
		}
	}
	return true
}

// getCachedPage returns the session's cached page.
func getCachedPage(sessionId string) string {
	for _, session := range sessions {
		if session.id == sessionId {
			return session.cachedPage
		}
	}
	return ""
}

