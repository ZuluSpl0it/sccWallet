package nix

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
	if browserCfg != "chromium" && chromium() {
		return true
	}
	if browserCfg != "firefox" && firefox() {
		return true
	}
	err := browser.OpenURL("http://localhost:4300")
	return err == nil
}

func edge() bool {
	err := exec.Command("microsoft-edge", "--app=http://localhost:4300").Run()
	if err == nil {
		return true
	}
	err = exec.Command("microsoft-edge-stable", "--app=http://localhost:4300").Run()
	if err == nil {
		return true
	}
	err = exec.Command("/usr/bin/microsoft-edge", "--app=http://localhost:4300").Run()
	if err == nil {
		return true
	}
	err = exec.Command("/usr/bin/microsoft-edge-stable", "--app=http://localhost:4300").Run()
	return err == nil
}

func chrome() bool {
	err := exec.Command("google-chrome", "--app=http://localhost:4300").Run()
	if err == nil {
		return true
	}
	err = exec.Command("google-chrome-stable", "--app=http://localhost:4300").Run()
	if err == nil {
		return true
	}
	err = exec.Command("/usr/bin/google-chrome", "--app=http://localhost:4300").Run()
	if err == nil {
		return true
	}
	err = exec.Command("/usr/bin/google-chrome-stable", "--app=http://localhost:4300").Run()
	return err == nil
}

func chromium() bool {
	err := exec.Command("chromium", "--app=http://localhost:4300").Run()
	if err == nil {
		return true
	}
	err = exec.Command("chromium-browser", "--app=http://localhost:4300").Run()
	if err == nil {
		return true
	}
	err = exec.Command("/usr/bin/chromium", "--app=http://localhost:4300").Run()
	if err == nil {
		return true
	}
	err = exec.Command("/usr/bin/chromium-browser", "--app=http://localhost:4300").Run()
	return err == nil
}

func firefox() bool {
	err := exec.Command("firefox", "--new-window", "http://localhost:4300").Run()
	if err == nil {
		return true
	}
	err = exec.Command("/usr/bin/firefox", "--new-window", "http://localhost:4300").Run()
	if err == nil {
		return true
	}
	return err == nil
}
