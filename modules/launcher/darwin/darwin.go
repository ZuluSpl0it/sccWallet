package darwin

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
	case "safari":
		return safari()
	}
	return false
}

func fallback(browserCfg string) bool {
	if browserCfg != "chrome" && chrome() {
		return true
	}
	if browserCfg != "edge" && edge() {
		return true
	}
	if browserCfg != "firefox" && firefox() {
		return true
	}
	if browserCfg != "safari" && safari() {
		return true
	}
	err := browser.OpenURL("http://localhost:4300")
	return err == nil
}

func edge() bool {
	err := exec.Command("open", "-n", "-a", "Microsoft Edge", "--args", "--app=http://localhost:4300").Run()
	return err == nil
}

func chrome() bool {
	err := exec.Command("open", "-n", "-a", "Google Chrome", "--args", "--app=http://localhost:4300").Run()
	return err == nil
}

func firefox() bool {
	err := exec.Command("open", "-n", "-a", "Firefox", "--args", "-new-window=http://localhost:4300").Run()
	return err == nil
}

func safari() bool {
	err := exec.Command("open", "-a", "Safari", "http://localhost:4300").Run()
	return err == nil
}
