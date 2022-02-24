package server

import (
	"net/http"

  "github.com/julienschmidt/httprouter"
)

func buildHttpRoutes() *httprouter.Router {
  router := httprouter.New()
  router.NotFound = http.HandlerFunc(notFoundHandler)
  router.RedirectTrailingSlash = false

  //GUI Calls
  router.GET("/favicon.ico", faviconHandler)
  router.GET("/gui/balance", balanceHandler)
  router.GET("/gui/blockHeight", blockHeightHandler)
  router.GET("/gui/downloaderProgress", downloaderProgressHandler)
  router.GET("/gui/heartbeat", heartbeatHandler)
  router.GET("/gui/logo.png", logoHandler)
  router.GET("/gui/scripts.js", scriptHandler)
  router.GET("/gui/styles.css", styleHandler)
  router.GET("/gui/fonts/open-sans-v27-latin-regular.woff2", openSansLatinRegularWoff2Handler)
  router.GET("/gui/fonts/open-sans-v27-latin-700.woff2", openSansLatin700Woff2Handler)
  if n == nil {
    router.GET("/", downloadingHandler)
  } else {
    router.GET("/", guiHandler)
	  router.GET("/gui", guiHandler)
	  router.GET("/gui/export", redirect)
	  router.GET("/gui/alert/changeLock", redirect)
	  router.GET("/gui/alert/initializeSeed", redirect)
	  router.GET("/gui/alert/sendCoins", redirect)
	  router.GET("/gui/alert/receiveCoins", redirect)
	  router.GET("/gui/alert/recoverSeed", redirect)
	  router.GET("/gui/alert/restoreFromSeed", redirect)
	  router.GET("/gui/changeLock", redirect)
	  router.GET("/gui/collapseMenu", redirect)
	  router.GET("/gui/expandMenu", redirect)
	  router.GET("/gui/initializeSeed", redirect)
	  router.GET("/gui/lockWallet", redirect)
	  router.GET("/gui/restoreSeed", redirect)
	  router.GET("/gui/scanning", redirect)
	  router.GET("/gui/sendCoins", redirect)
    router.GET("/gui/setTxHistoryPage", redirect)
	  router.GET("/gui/unlockWallet", redirect)
	  router.GET("/gui/explorer", redirect)
	  router.POST("/gui", guiHandler)
	  router.POST("/gui/export", transactionHistoryCsvExport)
	  router.POST("/gui/alert/changeLock", alertChangeLockHandler)
	  router.POST("/gui/alert/initializeSeed", alertInitializeSeedHandler)
	  router.POST("/gui/alert/sendCoins", alertSendCoinsHandler)
	  router.POST("/gui/alert/receiveCoins", alertReceiveCoinsHandler)
	  router.POST("/gui/alert/recoverSeed", alertRecoverSeedHandler)
	  router.POST("/gui/alert/restoreFromSeed", alertRestoreFromSeedHandler)
	  router.POST("/gui/changeLock", changeLockHandler)
	  router.POST("/gui/collapseMenu", collapseMenuHandler)
	  router.POST("/gui/expandMenu", expandMenuHandler)
	  router.POST("/gui/initializeSeed", initializeSeedHandler)
	  router.POST("/gui/lockWallet", lockWalletHandler)
	  router.POST("/gui/restoreSeed", restoreSeedHandler)
	  router.POST("/gui/scanning", scanningHandler)
	  router.POST("/gui/sendCoins", sendCoinsHandler)
	  router.POST("/gui/setTxHistoryPage", setTxHistoyPage)
	  router.POST("/gui/unlockWallet", unlockWalletHandler)
	  router.POST("/gui/explorer", explorerHandler)
  }
  return router
}
