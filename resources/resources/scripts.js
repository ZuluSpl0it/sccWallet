function refreshBlockHeight(sessionID) {
  if (document.getElementsByClassName('block_height').length > 0) {
    var data = new FormData();
    data.append("session_id", sessionID)
    fetch("/gui/blockHeight", {method: "POST", body: data})
      .then(response => response.json())
      .then(result => {
        var blockHeight = result[0]
        var status = result[1]
        var color = result[2]
        // Automatically refresh form to make GUI smoother.
        if (status === "Synchronized") {
          var refreshForm = document.getElementById("refreshForm")
          if (typeof(refreshForm) != 'undefined' && refreshForm != null) {
            refreshForm.submit()
          }
        }
        for (const element of document.getElementsByClassName("block_height")){
          element.innerHTML=blockHeight;
        }
        for (const element of document.getElementsByClassName("status")){
          element.innerHTML=status;
        }
        for (const element of document.getElementsByClassName("status")){
          element.className="status " + color
        }
        setTimeout(() => {refreshBlockHeight(sessionID);}, 1000);
      })
      .catch(error => {
        console.error("Error:", error);
        setTimeout(() => {refreshBlockHeight(sessionID);}, 1000);
      })
  } else {
    setTimeout(() => {refreshBlockHeight(sessionID);}, 50);
  }
}
function isLastPage() {
  var isLastPageElement = document.getElementById("is_last_page")
  if (typeof(isLastPageElement) != 'undefined' && isLastPageElement != null) {
    return isLastPageElement.className === "true";
  }
  return false;
}
function refreshBalance(sessionID) {
  var balance = document.getElementById("balance");
  if (typeof(balance) != 'undefined' && balance != null) {
    var data = new FormData();
    data.append("session_id", sessionID)
    fetch("/gui/balance", {method: "POST", body: data})
      .then(response => response.json())
      .then(result => {
        for (const element of document.getElementsByClassName("confirmed")){
          element.innerHTML = result[0];
        }
        for (const element of document.getElementsByClassName("unconfirmed")){
          if (isLastPage() && element.innerHTML.trim() !== result[1].trim()) {
            var refreshTransactions = document.getElementById("refresh_transactions")
            if (typeof(refreshTransactions) != 'undefined' && refreshTransactions != null) {
              refreshTransactions.submit()
            }
          }
          element.innerHTML = result[1];
        }
        for (const element of document.getElementsByClassName("spf_funds")){
          element.innerHTML = result[2];
        }
        var whaleSize = document.getElementById("whale_size")
        if (typeof(whaleSize) != 'undefined' && whaleSize != null) {
          whaleSize.innerHTML = "Whale Size: " + result[4];
        }
        var whaleSizeButton = document.getElementById("whale_size_button")
        if (typeof(whaleSizeButton) != 'undefined' && whaleSizeButton != null) {
          whaleSizeButton.value = "Whale Size: " + result[4];
        }
        setTimeout(() => {refreshBalance(sessionID);}, 1000);
      })
      .catch(error => {
        console.error("Error:", error);
        setTimeout(() => {refreshBalance(sessionID);}, 1000);
      })
  } else {
    setTimeout(() => {refreshBalance(sessionID);}, 50);
  }
}
function refreshBootstrapperProgress() {
  if (document.getElementsByClassName('bootstrapper-progress').length > 0) {
    fetch("/gui/bootstrapperProgress")
      .then(response => response.json())
      .then(result => {
        var status = result[0]
        // Autorefresh wallet to make onboarding smoother.
        if (status === "100%") {
          var refreshBootstrapper = document.getElementById("refreshBootstrapper")
          if (typeof(refreshBootstrapper) != 'undefined' && refreshBootstrapper != null) {
            refreshBootstrapper.submit()
          }
        }
        for (const element of document.getElementsByClassName("bootstrapper-progress")){
          element.innerHTML = status;
        }
        setTimeout(() => {refreshBootstrapperProgress();}, 1000);
      })
      .catch(error => {
        console.error("Error:", error);
        setTimeout(() => {refreshBootstrapperProgress();}, 1000);
      })
  } else {
    setTimeout(() => {refreshBootstrapperProgress();}, 50);
  }
}
function refreshConsensusBuilderProgress() {
  if (document.getElementsByClassName('consensus-builder-progress').length > 0) {
    fetch("/gui/consensusBuilderProgress")
      .then(response => response.json())
      .then(result => {
        var status = result[0]
        // Autorefresh wallet to make onboarding smoother.
        if (status === "100%") {
          var refreshConsensusBuilder = document.getElementById("refreshConsensusBuilder")
          if (typeof(refreshConsensusBuilder) != 'undefined' && refreshConsensusBuilder != null) {
            refreshConsensusBuilder.submit()
          }
        }
        for (const element of document.getElementsByClassName("consensus-builder-progress")){
          element.innerHTML = status;
        }
        setTimeout(() => {refreshConsensusBuilderProgress();}, 1000);
      })
      .catch(error => {
        console.error("Error:", error);
        setTimeout(() => {refreshConsensusBuilderProgress();}, 1000);
      })
  } else {
    setTimeout(() => {refreshConsensusBuilderProgress();}, 50);
  }
}
function refreshHeartbeat(sessionID) {
  var data = new FormData();
  data.append("session_id", sessionID)
  fetch("/gui/heartbeat", {method: "POST", body: data})
    .then(response => response.json())
    .then(result => {
    	if (result[0] === "true") {
        setTimeout(() => {refreshHeartbeat(sessionID);}, 200);
    	}
    })
    .catch(error => {
      console.error("Error:", error);
      shutdownNotice()
    })
}
function shutdownNotice() {
  document.body.innerHTML = `
    <div class="col-5 left top no-wrap">
      <div>
        <img class="scprime-logo" alt="ScPrime Web Wallet" src="/gui/logo.png"/>
      </div>
    </div>
    <div id="popup" class="popup center">
      <h2 class="uppercase">Shutdown Notice</h2>
      <div class="middle pad blue-dashed" id="popup_content">Wallet was shutdown.</div>
    </div>
    <div id="fade" class="fade"></div>
  `
}
function copyToClipboard(textToCopy) {
  var temp = document.createElement("input");
  temp.type = "text";
  temp.value = textToCopy;
  document.body.appendChild(temp);
  temp.select();
  document.execCommand("Copy");
  document.body.removeChild(temp);
}
refreshBootstrapperProgress()
refreshConsensusBuilderProgress()
refreshHeartbeat("")

