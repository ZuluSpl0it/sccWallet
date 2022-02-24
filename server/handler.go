package server

import (
  "context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"strconv"
	"strings"
	"time"

  "gitlab.com/scpcorp/webwallet/resources"

	"gitlab.com/NebulousLabs/errors"

	"gitlab.com/scpcorp/ScPrime/crypto"
	"gitlab.com/scpcorp/ScPrime/modules"
	"gitlab.com/scpcorp/ScPrime/modules/downloader"
	"gitlab.com/scpcorp/ScPrime/modules/wallet"
	"gitlab.com/scpcorp/ScPrime/types"

	"github.com/julienschmidt/httprouter"
	mnemonics "gitlab.com/NebulousLabs/entropy-mnemonics"
)

func notFoundHandler(w http.ResponseWriter, req *http.Request) {
  http.Error(w, "404 not found.", http.StatusNotFound)
}

func redirect(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	http.Redirect(w, req, "/", http.StatusMovedPermanently)
}

func faviconHandler(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	var favicon = resources.Favicon()
	w.Header().Set("Content-Type", "image/x-icon")
	w.Header().Set("Content-Length", strconv.Itoa(len(favicon))) //len(dec)
	w.Write(favicon)
}

func balanceHandler(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	fmtScpBal, fmtUncBal, fmtSpfBal, fmtClmBal := balancesHelper()
	writeArray(w, []string{fmtScpBal, fmtUncBal, fmtSpfBal, fmtClmBal})
}

func blockHeightHandler(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	fmtHeight, fmtStatus, fmtStatCo := blockHeightHelper()
	writeArray(w, []string{fmtHeight, fmtStatus, fmtStatCo})
}

func downloaderProgressHandler(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	writeArray(w, []string{downloader.Progress()})
}

func heartbeatHandler(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	updateHeartbeat()
	go shutdownHelper()
	writeArray(w, []string{"true"})
}

func logoHandler(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	var logo = resources.Logo()
	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Content-Length", strconv.Itoa(len(logo))) //len(dec)
	w.Write(logo)
}

func scriptHandler(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	var javascript = resources.Javascript()
	w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
	w.Header().Set("Content-Length", strconv.Itoa(len(javascript))) //len(dec)
	w.Write(javascript)
}

func styleHandler(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	var cssStyleSheet = resources.CssStyleSheet()
	w.Header().Set("Content-Type", "text/css; charset=utf-8")
	w.Header().Set("Content-Length", strconv.Itoa(len(cssStyleSheet))) //len(dec)
	w.Write(cssStyleSheet)
}

func openSansLatinRegularWoff2Handler(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	var font = resources.OpenSansLatinRegularWoff2()
	w.Header().Set("Content-Type", "font/woff2")
	w.Header().Set("Content-Length", strconv.Itoa(len(font))) //len(dec)
	w.Write(font)
}

func openSansLatin700Woff2Handler(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	var font = resources.OpenSansLatin700Woff2()
	w.Header().Set("Content-Type", "font/woff2")
	w.Header().Set("Content-Length", strconv.Itoa(len(font))) //len(dec)
	w.Write(font)
}

func transactionHistoryCsvExport(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	history, err := transctionHistoryCsvExportHelper()
	if err != nil {
		history = "failed"
	}
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-disposition", "attachment;filename=history.csv")
	w.Header().Set("Content-Length", strconv.Itoa(len(history))) //len(dec)
	w.Write([]byte(history))
}

func transctionHistoryCsvExportHelper() (string, error) {
	csv := `"Transaction ID","Type","Amount SCP","Amount SPF","Confirmed","DateTime"` + "\n"
	heightMin := 0
	confirmedTxns, err := n.Wallet.Transactions(types.BlockHeight(heightMin), n.ConsensusSet.Height())
	if err != nil {
		return "", err
	}
	unconfirmedTxns, err := n.Wallet.UnconfirmedTransactions()
	if err != nil {
		return "", err
	}
	sts, err := wallet.ComputeSummarizedTransactions(append(confirmedTxns, unconfirmedTxns...), n.ConsensusSet.Height())
	if err != nil {
		return "", err
	}
	for _, txn := range sts {
		// Format transaction type
		if txn.Type != "SETUP" {
			fmtSpf := txn.Spf
			if fmtSpf == "" {
				fmtSpf = "0"
			}
			csv = csv + fmt.Sprintf(`"%s","%s","%s","%s","%s","%s"`, txn.TxnId, txn.Type, txn.Scp, fmtSpf, txn.Confirmed, txn.Time) + "\n"
		}
	}
	return csv, nil
}

func alertChangeLockHandler(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	sessionId := req.FormValue("session_id")
	if sessionId == "" || !sessionIdExists(sessionId) {
		msg := "Session ID does not exist."
		writeError(w, msg, "")
	}
	title := "CHANGE LOCK"
	form := resources.ChangeLockForm()
	writeForm(w, title, form, sessionId)
}

func alertInitializeSeedHandler(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	title := "CREATE NEW WALLET"
	form := resources.IntializeSeedForm()
	writeForm(w, title, form, "")
}

func alertSendCoinsHandler(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	sessionId := req.FormValue("session_id")
	if sessionId == "" || !sessionIdExists(sessionId) {
		msg := "Session ID does not exist."
		writeError(w, msg, "")
	}
	title := "SEND"
	form := resources.SendCoinsForm()
	writeForm(w, title, form, sessionId)
}

func alertReceiveCoinsHandler(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	sessionId := req.FormValue("session_id")
	if sessionId == "" || !sessionIdExists(sessionId) {
		msg := "Session ID does not exist."
		writeError(w, msg, "")
	}
	var msgPrefix = "Unable to retrieve address: "
	addresses, err := n.Wallet.LastAddresses(1)
	if err != nil {
		msg := fmt.Sprintf("%s%v", msgPrefix, err)
		writeError(w, msg, sessionId)
		return
	}
	if len(addresses) == 0 {
		_, err := n.Wallet.NextAddress()
		if err != nil {
			msg := fmt.Sprintf("%s%v", msgPrefix, err)
			writeError(w, msg, sessionId)
			return
		}
	}
	addresses, err = n.Wallet.LastAddresses(1)
	if err != nil {
		msg := fmt.Sprintf("%s%v", msgPrefix, err)
		writeError(w, msg, sessionId)
		return
	}
	title := "RECEIVE"
	msg := strings.ToUpper(fmt.Sprintf("%s", addresses[0]))
	writeMsg(w, title, msg, sessionId)
}

func alertRecoverSeedHandler(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	sessionId := req.FormValue("session_id")
	if sessionId == "" || !sessionIdExists(sessionId) {
		msg := "Session ID does not exist."
		writeError(w, msg, "")
	}
	cancel := req.FormValue("cancel")
	var msgPrefix = "Unable to recover seed: "
	if cancel == "true" {
		guiHandler(w, req, nil)
		return
	}
	unlocked, err := n.Wallet.Unlocked()
	if err != nil {
		msg := fmt.Sprintf("%s%v", msgPrefix, err)
		writeError(w, msg, "")
		return
	}
	if !unlocked {
		msg := msgPrefix + "Wallet is locked."
		writeError(w, msg, "")
		return
	}
	// Get the primary seed information.
	dictionary := mnemonics.DictionaryID(req.FormValue("dictionary"))
	if dictionary == "" {
		dictionary = mnemonics.English
	}
	primarySeed, _, err := n.Wallet.PrimarySeed()
	if err != nil {
		msg := fmt.Sprintf("%s%v", msgPrefix, err)
		writeError(w, msg, sessionId)
		return
	}
	primarySeedStr, err := modules.SeedToString(primarySeed, dictionary)
	if err != nil {
		msg := fmt.Sprintf("%s%v", msgPrefix, err)
		writeError(w, msg, sessionId)
		return
	}
	title := "RECOVER SEED"
	msg := fmt.Sprintf("%s", primarySeedStr)
	writeMsg(w, title, msg, sessionId)
}

func alertRestoreFromSeedHandler(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	if !n.ConsensusSet.Synced() {
		msg := "Wallet must be syncronized with the network before it can be restored from a seed."
		writeError(w, msg, "")
		return
	}
	title := "RESTORE FROM SEED"
	form := resources.RestoreFromSeedForm()
	writeForm(w, title, form, "")
}

func changeLockHandler(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	sessionId := req.FormValue("session_id")
	if sessionId == "" || !sessionIdExists(sessionId) {
		msg := "Session ID does not exist."
		writeError(w, msg, "")
	}
	cancel := req.FormValue("cancel")
	origPassword := req.FormValue("orig_password")
	newPassword := req.FormValue("new_password")
	confirmPassword := req.FormValue("confirm_password")
	var msgPrefix = "Unable to change lock: "
	if cancel == "true" {
		guiHandler(w, req, nil)
		return
	}
	if origPassword == "" {
		msg := msgPrefix + "The original password must be provided."
		writeError(w, msg, sessionId)
		return
	}
	validPass, err := isPasswordValid(origPassword)
	if err != nil {
		msg := fmt.Sprintf("%s%v", msgPrefix, err)
		writeError(w, msg, sessionId)
		return
	} else if !validPass {
		msg := msgPrefix + "The original password is not valid."
		writeError(w, msg, sessionId)
		return
	}
	if newPassword == "" {
		msg := msgPrefix + "A new password must be provided."
		writeError(w, msg, sessionId)
		return
	}
	if confirmPassword == "" {
		msg := msgPrefix + "A confirmation password must be provided."
		writeError(w, msg, sessionId)
		return
	}
	if newPassword != confirmPassword {
		msg := msgPrefix + "New password does not match confirmation password."
		writeError(w, msg, sessionId)
		return
	}
	var newKey crypto.CipherKey
	newKey = crypto.NewWalletKey(crypto.HashObject(newPassword))
	primarySeed, _, _ := n.Wallet.PrimarySeed()
	err = n.Wallet.ChangeKeyWithSeed(primarySeed, newKey)
	if err != nil {
		msg := fmt.Sprintf("%s%v", msgPrefix, err)
		writeError(w, msg, sessionId)
		return
	}
	guiHandler(w, req, nil)
}

func initializeSeedHandler(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	cancel := req.FormValue("cancel")
	newPassword := req.FormValue("new_password")
	confirmPassword := req.FormValue("confirm_password")
	var msgPrefix = "Unable to initialize new wallet seed: "
	if cancel == "true" {
		guiHandler(w, req, nil)
		return
	}
	if newPassword == "" {
		msg := msgPrefix + "A new password must be provided."
		writeError(w, msg, "")
		return
	}
	if confirmPassword == "" {
		msg := msgPrefix + "A confirmation password must be provided."
		writeError(w, msg, "")
		return
	}
	if newPassword != confirmPassword {
		msg := msgPrefix + "New password does not match confirmation password."
		writeError(w, msg, "")
		return
	}
	encrypted, err := n.Wallet.Encrypted()
	if err != nil {
		msg := fmt.Sprintf("%s%v", msgPrefix, err)
		writeError(w, msg, "")
		return
	}
	if encrypted {
		msg := msgPrefix + "Seed was already initialized."
		writeError(w, msg, "")
		return
	}
	go initializeSeedHelper(newPassword)
	title := "<font class='status &STATUS_COLOR;'>&STATUS;</font> WALLET"
	form := resources.ScanningWalletForm()
	sessionId := addSessionId()
	writeForm(w, title, form, sessionId)
}

func lockWalletHandler(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	sessionId := req.FormValue("session_id")
	if sessionId == "" || !sessionIdExists(sessionId) {
		msg := "Session ID does not exist."
		writeError(w, msg, "")
	}
	cancel := req.FormValue("cancel")
	var msgPrefix = "Unable to lock wallet: "
	if cancel == "true" {
		guiHandler(w, req, nil)
		return
	}
	unlocked, err := n.Wallet.Unlocked()
	if err != nil {
		msg := fmt.Sprintf("%s%v", msgPrefix, err)
		writeError(w, msg, "")
		return
	}
	if !unlocked {
		msg := msgPrefix + "Wallet was already locked."
		writeError(w, msg, "")
		return
	}
	n.Wallet.Lock()
	guiHandler(w, req, nil)
}

func restoreSeedHandler(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	cancel := req.FormValue("cancel")
	newPassword := req.FormValue("new_password")
	confirmPassword := req.FormValue("confirm_password")
	seedStr := req.FormValue("seed_str")
	var msgPrefix = "Unable to restore wallet from seed: "
	if cancel == "true" {
		guiHandler(w, req, nil)
		return
	}
	if newPassword == "" {
		msg := msgPrefix + "A new password must be provided."
		writeError(w, msg, "")
		return
	}
	if confirmPassword == "" {
		msg := msgPrefix + "A confirmation password must be provided."
		writeError(w, msg, "")
		return
	}
	if newPassword != confirmPassword {
		msg := msgPrefix + "New password does not match confirmation password."
		writeError(w, msg, "")
		return
	}
	if seedStr == "" {
		msg := msgPrefix + "A seed must be provided."
		writeError(w, msg, "")
		return
	}
	encrypted, err := n.Wallet.Encrypted()
	if err != nil {
		msg := fmt.Sprintf("%s%v", msgPrefix, err)
		writeError(w, msg, "")
		return
	}
	if encrypted {
		msg := msgPrefix + "Seed is already initialized."
		writeError(w, msg, "")
		return
	}
	seed, err := modules.StringToSeed(seedStr, "english")
	if err != nil {
		msg := fmt.Sprintf("%s%v", msgPrefix, err)
		writeError(w, msg, "")
		return
	}
	go restoreSeedHelper(newPassword, seed)
	title := "<font class='status &STATUS_COLOR;'>&STATUS;</font> WALLET"
	form := resources.ScanningWalletForm()
	sessionId := addSessionId()
	writeForm(w, title, form, sessionId)
}

func sendCoinsHandler(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	sessionId := req.FormValue("session_id")
	if sessionId == "" || !sessionIdExists(sessionId) {
		msg := "Session ID does not exist."
		writeError(w, msg, "")
	}
	cancel := req.FormValue("cancel")
	var msgPrefix = "Unable to send coins: "
	if cancel == "true" {
		guiHandler(w, req, nil)
		return
	}
	unlocked, err := n.Wallet.Unlocked()
	if err != nil {
		msg := fmt.Sprintf("%s%v", msgPrefix, err)
		writeError(w, msg, "")
		return
	}
	if !unlocked {
		msg := msgPrefix + "Wallet is locked."
		writeError(w, msg, "")
		return
	}
	// Verify destination address was supplied.
	dest, err := scanAddress(req.FormValue("destination"))
	if err != nil {
		msg := msgPrefix + "Destination is not valid."
		writeError(w, msg, sessionId)
		return
	}
	coinType := req.FormValue("coin_type")
	if coinType == "SCP" {
		amount, err := types.NewCurrencyStr(req.FormValue("amount") + "SCP")
		if err != nil {
			msg := fmt.Sprintf("%s%v", msgPrefix, err)
			writeError(w, msg, sessionId)
			return
		}
		_, err = n.Wallet.SendSiacoins(amount, dest)
		if err != nil {
			msg := fmt.Sprintf("%s%v", msgPrefix, err)
			writeError(w, msg, sessionId)
			return
		}
	} else if coinType == "SPF" {
		amount, err := types.NewCurrencyStr(req.FormValue("amount") + "SPF")
		if err != nil {
			msg := fmt.Sprintf("%s%v", msgPrefix, err)
			writeError(w, msg, sessionId)
			return
		}
		_, err = n.Wallet.SendSiafunds(amount, dest)
		if err != nil {
			msg := fmt.Sprintf("%s%v", msgPrefix, err)
			writeError(w, msg, sessionId)
			return
		}
	} else {
		msg := msgPrefix + "Coin type was not supplied."
		writeError(w, msg, sessionId)
		return
	}
	guiHandler(w, req, nil)
}

func unlockWalletHandler(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	password := req.FormValue("password")
	var msgPrefix = "Unable to unlock wallet: "
	if password == "" {
		msg := "A password must be provided."
		writeError(w, msgPrefix+msg, "")
		return
	}
	potentialKeys, _ := encryptionKeys(password)
	for _, key := range potentialKeys {
		unlocked, err := n.Wallet.Unlocked()
		if err != nil {
			msg := fmt.Sprintf("%s%v", msgPrefix, err)
			writeError(w, msg, "")
			return
		}
		if !unlocked {
			n.Wallet.Unlock(key)
		}
	}
	unlocked, err := n.Wallet.Unlocked()
	if err != nil {
		msg := fmt.Sprintf("%s%v", msgPrefix, err)
		writeError(w, msg, "")
		return
	}
	if !unlocked {
		msg := msgPrefix + "Password is not valid."
		writeError(w, msg, "")
		return
	}
	sessionId := addSessionId()
	if status != "" {
		title := "<font class='status &STATUS_COLOR;'>&STATUS;</font> WALLET"
		form := resources.ScanningWalletForm()
		writeForm(w, title, form, sessionId)
		return
	}
	writeWallet(w, sessionId)
}

func explorerHandler(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	sessionId := req.FormValue("session_id")
	if sessionId == "" || !sessionIdExists(sessionId) {
		form := resources.UnlockWalletForm()
		writeForm(w, "UNLOCK WALLET", form, "")
		return
	}
	var msgPrefix = "Unable to retrieve the transaction: "
	if req.FormValue("transaction_id") == "" {
		msg := msgPrefix + "No transaction ID was provided."
		writeError(w, msg, sessionId)
		return
	}
	var transactionId types.TransactionID
	jsonID := "\"" + req.FormValue("transaction_id") + "\""
	err := transactionId.UnmarshalJSON([]byte(jsonID))
	if err != nil {
		msg := msgPrefix + "Unable to parse transaction ID."
	  writeError(w, msg, sessionId)
		return
	}
	txn, ok, err := n.Wallet.Transaction(transactionId)
	if err != nil {
		msg := fmt.Sprintf("%s%v", msgPrefix, err)
		writeError(w, msg, sessionId)
		return
	}
	if !ok {
		msg := msgPrefix + "Transaction was not found."
		writeError(w, msg, sessionId)
		return
	}
	transactionDetails, _ := transactionExplorerHelper(txn)
	html := resources.WalletHtmlTemplate()
	html = strings.Replace(html, "&TRANSACTION_PORTAL;", transactionDetails, -1)
	writeHtml(w, html, sessionId)
}

func downloadingHandler(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	html := strings.Replace(resources.DownloadingHtml(), "&DOWNLOADER_PROGRESS;", downloader.Progress(), -1)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, html)
}

func loadingHandler(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	var body = resources.LoadingHtml()
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Content-Length", strconv.Itoa(len(body))) //len(dec)
	w.Write(body)
}

func notLoadedHandler(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	writeArray(w, []string{"The GUI module is not loaded."})
}

func expandMenuHandler(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	sessionId := req.FormValue("session_id")
	if sessionId == "" || !sessionIdExists(sessionId) {
		form := resources.UnlockWalletForm()
		writeForm(w, "UNLOCK WALLET", form, "")
		return
	}
	expandMenu(sessionId)
	writeHtml(w, getCachedPage(sessionId), sessionId)
}

func collapseMenuHandler(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	sessionId := req.FormValue("session_id")
	if sessionId == "" || !sessionIdExists(sessionId) {
		form := resources.UnlockWalletForm()
		writeForm(w, "UNLOCK WALLET", form, "")
		return
	}
	collapseMenu(sessionId)
	writeHtml(w, getCachedPage(sessionId), sessionId)
}

func scanningHandler(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	sessionId := req.FormValue("session_id")
	height, _, _ := blockHeightHelper()
	if height == "0" && status != "" {
		title := "<font class='status &STATUS_COLOR;'>&STATUS;</font> WALLET"
		form := resources.ScanningWalletForm()
		writeForm(w, title, form, sessionId)
		return
	}
	if sessionId == "" || !sessionIdExists(sessionId) {
		form := resources.UnlockWalletForm()
		writeForm(w, "UNLOCK WALLET", form, "")
		return
	}
	if status != "" {
		title := "<font class='status &STATUS_COLOR;'>&STATUS;</font> WALLET"
		form := resources.ScanningWalletForm()
		writeForm(w, title, form, sessionId)
		return
	}
	guiHandler(w, req, nil)
}

func setTxHistoyPage(w http.ResponseWriter, req *http.Request, resp httprouter.Params) {
	sessionId := req.FormValue("session_id")
	if sessionId == "" || !sessionIdExists(sessionId) {
		form := resources.UnlockWalletForm()
		writeForm(w, "UNLOCK WALLET", form, "")
		return
	}
	page, _ := strconv.Atoi(req.FormValue("page"))
	setTxHistoryPage(page, sessionId)
	guiHandler(w, req, nil)
}

func guiHandler(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	sessionId := req.FormValue("session_id")
	height, _, _ := blockHeightHelper()
	if height == "0" && status != "" {
		title := "<font class='status &STATUS_COLOR;'>&STATUS;</font> WALLET"
		form := resources.ScanningWalletForm()
		writeForm(w, title, form, sessionId)
		return
	}
	encrypted, err := n.Wallet.Encrypted()
	if err != nil {
		msg := fmt.Sprintf("Unable to determine if wallet is encrypted: %v", err)
		writeError(w, msg, sessionId)
		return
	}
	if !encrypted {
		title := "INITIALIZE WALLET"
		form := resources.InitializeWalletForm()
		writeForm(w, title, form, sessionId)
		return
	}
	if sessionId == "" || !sessionIdExists(sessionId) {
		form := resources.UnlockWalletForm()
		writeForm(w, "UNLOCK WALLET", form, sessionId)
		return
	}
	if status != "" {
		title := "<font class='status &STATUS_COLOR;'>&STATUS;</font> WALLET"
		form := resources.ScanningWalletForm()
		writeForm(w, title, form, sessionId)
		return
	}
	unlocked, err := n.Wallet.Unlocked()
	if err != nil {
		msg := fmt.Sprintf("Unable to determine if wallet is unlocked: %v", err)
		writeError(w, msg, sessionId)
		return
	}
	if unlocked {
		writeWallet(w, sessionId)
		return
	}
	title := "UNLOCK WALLET"
	form := resources.UnlockWalletForm()
	writeForm(w, title, form, "")
}

func writeWallet(w http.ResponseWriter, sessionId string) {
	transactionHistoryLines, pages, err := transactionHistoryHelper(sessionId)
	if err != nil {
		msg := fmt.Sprintf("Unable to generate transaction history: %v", err)
		writeError(w, msg, sessionId)
		return
	}
	html := resources.WalletHtmlTemplate()
	html = strings.Replace(html, "&TRANSACTION_PORTAL;", resources.TransactionsHistoryHtmlTemplate(), -1)
	html = strings.Replace(html, "&TRANSACTION_HISTORY_LINES;", transactionHistoryLines, -1)
	options := ""
	for i := 0; i < pages+1; i++ {
		selected := ""
		if i+1 == getTxHistoryPage(sessionId) {
			selected = "selected"
		}
		options = fmt.Sprintf("<option %s value='%d'>%d</option>", selected, i+1, i+1) + options
	}
	if pages == 0 {
		html = strings.Replace(html, "&TRANSACTION_PAGINATION;", "<div class='col-4 center no-wrap'></div>", -1)
	} else {
		html = strings.Replace(html, "&TRANSACTION_PAGINATION;", resources.TransactionPaginationTemplate(), -1)
	}
	html = strings.Replace(html, "&TRANSACTION_HISTORY_PAGE;", options, -1)
	html = strings.Replace(html, "&TRANSACTION_HISTORY_PAGES;", strconv.Itoa(pages+1), -1)
	writeHtml(w, html, sessionId)
}

func writeArray(w http.ResponseWriter, arr []string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	encjson, _ := json.Marshal(arr)
	fmt.Fprint(w, string(encjson))
}

func writeError(w http.ResponseWriter, msg string, sessionId string) {
	html := resources.AlertHtmlTemplate()
	html = strings.Replace(html, "&POPUP_TITLE;", "ERROR", -1)
	html = strings.Replace(html, "&POPUP_CONTENT;", msg, -1)
	html = strings.Replace(html, "&POPUP_CLOSE;", resources.CloseAlertForm(), -1)
	fmt.Println(msg)
	writeHtml(w, html, sessionId)
}

func writeMsg(w http.ResponseWriter, title string, msg string, sessionId string) {
	html := resources.AlertHtmlTemplate()
	html = strings.Replace(html, "&POPUP_TITLE;", title, -1)
	html = strings.Replace(html, "&POPUP_CONTENT;", msg, -1)
	html = strings.Replace(html, "&POPUP_CLOSE;", resources.CloseAlertForm(), -1)
	writeHtml(w, html, sessionId)
}

func writeForm(w http.ResponseWriter, title string, form string, sessionId string) {
	html := resources.AlertHtmlTemplate()
	html = strings.Replace(html, "&POPUP_TITLE;", title, -1)
	html = strings.Replace(html, "&POPUP_CONTENT;", form, -1)
	html = strings.Replace(html, "&POPUP_CLOSE;", "", -1)
	writeHtml(w, html, sessionId)
}

func writeHtml(w http.ResponseWriter, html string, sessionId string) {
	cachedPage(html, sessionId)
	fmtHeight, fmtStatus, fmtStatCo := blockHeightHelper()
	html = strings.Replace(html, "&STATUS_COLOR;", fmtStatCo, -1)
	html = strings.Replace(html, "&STATUS;", fmtStatus, -1)
	html = strings.Replace(html, "&BLOCK_HEIGHT;", fmtHeight, -1)
	fmtScpBal, fmtUncBal, fmtSpfBal, fmtClmBal := balancesHelper()
	html = strings.Replace(html, "&SCP_BALANCE;", fmtScpBal, -1)
	html = strings.Replace(html, "&UNCONFIRMED_DELTA;", fmtUncBal, -1)
	html = strings.Replace(html, "&SPF_BALANCE;", fmtSpfBal, -1)
	html = strings.Replace(html, "&SCP_CLAIM_BALANCE;", fmtClmBal, -1)
	if menuIsCollapsed(sessionId) {
		html = strings.Replace(html, "&MENU;", resources.CollapsedMenuForm(), -1)
	} else {
		html = strings.Replace(html, "&MENU;", resources.ExpandedMenuForm(), -1)
	}
	html = strings.Replace(html, "&SESSION_ID;", sessionId, -1)
	// add random data to links to act as a cache buster.
	// must be done last in case a cache buster is added in from a template.
	b := make([]byte, 16) //32 characters long
	rand.Read(b)
	cacheBuster := hex.EncodeToString(b)
	html = strings.Replace(html, "&CACHE_BUSTER;", cacheBuster, -1)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, html)
}

func balancesHelper() (string, string, string, string) {
	unlocked, err := n.Wallet.Unlocked()
	if err != nil {
		fmt.Printf("Unable to determine if wallet is unlocked: %v", err)
	}
	fmtScpBal := "?"
	fmtUncBal := "?"
	fmtSpfBal := "?"
	fmtClmBal := "?"
	if unlocked {
		scpBal, spfBal, scpClaimBal, err := n.Wallet.ConfirmedBalance()
		if err != nil {
			fmt.Printf("Unable to obtain confirmed balance: %v", err)
		} else {
			scpBalFloat, _ := new(big.Rat).SetFrac(scpBal.Big(), types.ScPrimecoinPrecision.Big()).Float64()
			scpClaimBalFloat, _ := new(big.Rat).SetFrac(scpClaimBal.Big(), types.ScPrimecoinPrecision.Big()).Float64()
			fmtScpBal = fmt.Sprintf("%15.2f", scpBalFloat)
			fmtSpfBal = fmt.Sprintf("%s", spfBal)
			fmtClmBal = fmt.Sprintf("%15.2f", scpClaimBalFloat)
		}
		scpOut, scpIn, err := n.Wallet.UnconfirmedBalance()
		if err != nil {
			fmt.Printf("Unable to obtain unconfirmed balance: %v", err)
		} else {
			scpInFloat, _ := new(big.Rat).SetFrac(scpIn.Big(), types.ScPrimecoinPrecision.Big()).Float64()
			scpOutFloat, _ := new(big.Rat).SetFrac(scpOut.Big(), types.ScPrimecoinPrecision.Big()).Float64()
			fmtUncBal = fmt.Sprintf("%15.2f", (scpInFloat - scpOutFloat))
		}
	}
	return fmtScpBal, fmtUncBal, fmtSpfBal, fmtClmBal
}

func blockHeightHelper() (string, string, string) {
	fmtHeight := "?"
	height, err := n.Wallet.Height()
	if err != nil {
		fmt.Printf("Unable to obtain block height: %v", err)
	} else {
		fmtHeight = fmt.Sprintf("%d", height)
	}
	rescanning, err := n.Wallet.Rescanning()
	if err != nil {
		fmt.Printf("Unable to determine if wallet is being scanned: %v", err)
	}
  synced := n.ConsensusSet.Synced()
	if status != "" {
		return fmtHeight, status, "yellow"
	} else if rescanning {
		return fmtHeight, "Rescanning", "cyan"
	} else if synced {
		return fmtHeight, "Synchronized", "blue"
	} else {
		return fmtHeight, "Synchronizing", "yellow"
	}
}

func initializeSeedHelper(newPassword string) {
	setStatus("Initializing")
	var encryptionKey crypto.CipherKey = crypto.NewWalletKey(crypto.HashObject(newPassword))
	_, err := n.Wallet.Encrypt(encryptionKey)
	if err != nil {
		fmt.Printf("Unable to initialize new wallet seed: %v", err)
		return
	}
	potentialKeys, _ := encryptionKeys(newPassword)
	for _, key := range potentialKeys {
		unlocked, err := n.Wallet.Unlocked()
		if err != nil {
			fmt.Printf("Unable to initialize new wallet seed: %v", err)
			return
		}
		if !unlocked {
			n.Wallet.Unlock(key)
		}
	}
	setStatus("")
}

func isPasswordValid(password string) (bool, error) {
	keys, _ := encryptionKeys(password)
	var err error
	for _, key := range keys {
		valid, keyErr := n.Wallet.IsMasterKey(key)
		if keyErr == nil {
			if valid {
				return true, nil
			}
			return false, nil
		}
		err = errors.Compose(err, keyErr)
	}
	return false, err
}

func restoreSeedHelper(newPassword string, seed modules.Seed) {
	setStatus("Restoring")
	var encryptionKey crypto.CipherKey = crypto.NewWalletKey(crypto.HashObject(newPassword))
	err := n.Wallet.InitFromSeed(encryptionKey, seed)
	if err != nil {
		fmt.Printf("Unable to restore wallet seed: %v", err)
		return
	}
	potentialKeys, _ := encryptionKeys(newPassword)
	for _, key := range potentialKeys {
		unlocked, err := n.Wallet.Unlocked()
		if err != nil {
			fmt.Printf("Unable to initialize new wallet seed: %v", err)
			return
		}
		if !unlocked {
			n.Wallet.Unlock(key)
		}
	}
	setStatus("")
}

func shutdownHelper() {
	time.Sleep(5000 * time.Millisecond)
	if time.Now().After(heartbeat.Add(5000 * time.Millisecond)) {
		fmt.Println("Shutting Down...")
		srv.Shutdown(context.Background())
	}
}

func transactionExplorerHelper(txn modules.ProcessedTransaction) (string, error) {
	unixTime, _ := strconv.ParseInt(fmt.Sprintf("%v", txn.ConfirmationTimestamp), 10, 64)
	fmtTime := strings.ToUpper(time.Unix(unixTime, 0).Format("2006-01-02 15:04"))
	fmtTxnId := strings.ToUpper(fmt.Sprintf("%v", txn.TransactionID))
	fmtTxnType := strings.ToUpper(strings.Replace(fmt.Sprintf("%v", txn.TxType), "_", " ", -1))
	fmtTxnBlock := strings.ToUpper(fmt.Sprintf("%v", txn.ConfirmationHeight))
	html := resources.TransactionInfoTemplate()
	html = strings.Replace(html, "&TXN_TYPE;", fmtTxnType, -1)
	html = strings.Replace(html, "&TXN_ID;", fmtTxnId, -1)
	html = strings.Replace(html, "&TXN_TIME;", fmtTime, -1)
	html = strings.Replace(html, "&TXN_BLOCK;", fmtTxnBlock, -1)
	inputs := ""
	for _, input := range txn.Inputs {
		fmtValue := strings.ToUpper(fmt.Sprintf("%v", input.Value))
		fmtAddress := strings.ToUpper(fmt.Sprintf("%v", input.RelatedAddress))
		fmtFundType := strings.ToUpper(strings.Replace(fmt.Sprintf("%v", input.FundType), "_", " ", -1))
		fmtFundType = strings.Replace(fmtFundType, "SIACOIN", "SCP", -1)
		fmtFundType = strings.Replace(fmtFundType, "SIAFUND", "SPF", -1)
		row := resources.TransactionInputTemplate()
		row = strings.Replace(row, "&VALUE;", fmtValue, -1)
		row = strings.Replace(row, "&ADDRESS;", fmtAddress, -1)
		row = strings.Replace(row, "&FUND_TYPE;", fmtFundType, -1)
		inputs = inputs + row
	}
	html = strings.Replace(html, "&TXN_INPUTS;", inputs, -1)
	outputs := ""
	for _, output := range txn.Outputs {
		fmtValue := strings.ToUpper(fmt.Sprintf("%v", output.Value))
		fmtAddress := strings.ToUpper(fmt.Sprintf("%v", output.RelatedAddress))
		fmtFundType := strings.ToUpper(strings.Replace(fmt.Sprintf("%v", output.FundType), "_", " ", -1))
		fmtFundType = strings.Replace(fmtFundType, "SIACOIN", "SCP", -1)
		fmtFundType = strings.Replace(fmtFundType, "SIAFUND", "SPF", -1)
		row := resources.TransactionOutputTemplate()
		row = strings.Replace(row, "&VALUE;", fmtValue, -1)
		row = strings.Replace(row, "&ADDRESS;", fmtAddress, -1)
		row = strings.Replace(row, "&FUND_TYPE;", fmtFundType, -1)
		outputs = outputs + row
	}
	html = strings.Replace(html, "&TXN_OUTPUTS;", outputs, -1)
	return html, nil
}

func transactionHistoryHelper(sessionId string) (string, int, error) {
	html := ""
	page := getTxHistoryPage(sessionId)
	pageSize := 20
	pageMin := (page - 1) * pageSize
	pageMax := page * pageSize
	count := 0
	heightMin := 0
	confirmedTxns, err := n.Wallet.Transactions(types.BlockHeight(heightMin), n.ConsensusSet.Height())
	if err != nil {
		return "", -1, err
	}
	unconfirmedTxns, err := n.Wallet.UnconfirmedTransactions()
	if err != nil {
		return "", -1, err
	}
	sts, err := wallet.ComputeSummarizedTransactions(append(confirmedTxns, unconfirmedTxns...), n.ConsensusSet.Height())
	if err != nil {
		return "", -1, err
	}
	for _, txn := range sts {
		// Format transaction type
		isSetup := txn.Type == "SETUP" && txn.Scp == fmt.Sprintf("%15.2f SCP", float64(0))
		if !isSetup {
			count++
			if count >= pageMin && count < pageMax {
				fmtAmount := txn.Scp
				if txn.Spf != "" {
					fmtAmount = fmtAmount + "; " + txn.Spf
				}
				row := resources.TransactionHistoryLineHtmlTemplate()
				row = strings.Replace(row, "&TRANSACTION_ID;", txn.TxnId, -1)
				row = strings.Replace(row, "&TYPE;", txn.Type, -1)
				row = strings.Replace(row, "&TIME;", txn.Time, -1)
				row = strings.Replace(row, "&AMOUNT;", fmtAmount, -1)
				row = strings.Replace(row, "&CONFIRMED;", txn.Confirmed, -1)
				html = html + row
			}
		}
	}
	return html, count / pageSize, nil
}

// scanAddress scans a types.UnlockHash from a string.
// copied from "gitlab.com/scpcorp/ScPrime/node/scan.go"
func scanAddress(addrStr string) (addr types.UnlockHash, err error) {
  err = addr.LoadString(addrStr)
  if err != nil {
    return types.UnlockHash{}, err
  }
  return addr, nil
}

// encryptionKeys enumerates the possible encryption keys that can be derived
// from an input string.
// copied from "gitlab.com/scpcorp/ScPrime/node/wallet.go"
func encryptionKeys(seedStr string) (validKeys []crypto.CipherKey, seeds []modules.Seed) {
  dicts := []mnemonics.DictionaryID{"english", "german", "japanese"}
  for _, dict := range dicts {
    seed, err := modules.StringToSeed(seedStr, dict)
    if err != nil {
      continue
    }
    validKeys = append(validKeys, crypto.NewWalletKey(crypto.HashObject(seed)))
    seeds = append(seeds, seed)
  }
  validKeys = append(validKeys, crypto.NewWalletKey(crypto.HashObject(seedStr)))
  return
}
