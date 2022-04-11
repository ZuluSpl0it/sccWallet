package windows

import (
	"os/exec"

	"github.com/pkg/browser"
)

// Launch launches the browser
func Launch(browserCfg string) bool {
	if launch(browserCfg) {
		return true
	}
	return fallback(browserCfg)
}

func launch(browserCfg string) bool {
	switch browserCfg {
	case "edge":
		return edge()
	case "chrome":
		return chrome()
	case "firefox":
		return firefox()
	}
	return false
}

func fallback(browserCfg string) bool {
	if browserCfg != "edge" && edge() {
		return true
	}
	if browserCfg != "chrome" && chrome() {
		return true
	}
	if browserCfg != "firefox" && firefox() {
		return true
	}
	err := browser.OpenURL("http://localhost:4300")
	return err == nil
}

func edge() bool {
	err := exec.Command("C:/Program Files/Microsoft/Edge/Application/msedge.exe", "--app=http://localhost:4300").Run()
	if err == nil {
		return true
	}
	err = exec.Command("C:/Program Files (x86)/Microsoft/Edge/Application/msedge.exe", "--app=http://localhost:4300").Run()
	return err == nil
}

func chrome() bool {
	err := exec.Command("C:/Program Files/Google/Chrome/Application/chrome.exe", "--app=http://localhost:4300").Run()
	if err == nil {
		return true
	}
	err = exec.Command("C:/Program Files (x86)/Google/Chrome/Application/chrome.exe", "--app=http://localhost:4300").Run()
	return err == nil
}

func firefox() bool {
	err := exec.Command("C:/Program Files/Mozilla Firefox/firefox.exe", "-new-window", "http://localhost:4300").Run()
	if err == nil {
		return true
	}
	err = exec.Command("C:/Program Files (x86)/Mozilla Firefox/firefox.exe", "-new-window", "http://localhost:4300").Run()
	return err == nil
}
