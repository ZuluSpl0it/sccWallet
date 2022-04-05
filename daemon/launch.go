package daemon

import (
	"os/exec"
	"runtime"

	"github.com/pkg/browser"
)

// Browsers that are based on Chromium (such as Google Chrome and Microsoft Edge) are most
// desirable because they can be launched in app mode (which means that there is no address bar).
// This allows the GUI head feel most like a native application.
func launch() bool {
	switch runtime.GOOS {
	case "android":
		return android()
	case "darwin":
		return darwin()
	case "windows":
		return windows()
	}
	return nix()
}

func android() bool {
	err := exec.Command("am", "start", "--user", "0", "-a", "android.intent.action.VIEW", "-d", "http://localhost:4300").Run()
	return err == nil
}

func darwin() bool {
	err := exec.Command("open", "-n", "-a", "Google Chrome", "--args", "--app=http://localhost:4300").Run()
	if err == nil {
		return true
	}
	err = exec.Command("open", "-n", "-a", "Microsoft Edge", "--args", "--app=http://localhost:4300").Run()
	if err == nil {
		return true
	}
	// Firefox can no longer be launched without an address bar as the ssb parameter (site
	// specific browser) was removed in 2021.
	err = exec.Command("open", "-n", "-a", "Firefox", "--args", "-new-window=http://localhost:4300").Run()
	if err == nil {
		return true
	}
	// Use Safari as a last resort as if it is not the default browser then two GUI heads will
	// open (one in Safari and one in the default browser).
	err = exec.Command("open", "-a", "Safari", "http://localhost:4300").Run()
	if err == nil {
		return true
	}
	err = browser.OpenURL("http://localhost:4300")
	return err == nil
}

func windows() bool {
	// Edge is installed on all Windows OSes and can not be removed. It also can, like chrome, be
	// launched in app mode which makes the experience feel like a native app. Thus, on Windows,
	// lets default to launching Edge and only fall back to chrome on older versions of windows.
	err := exec.Command("C:/Program Files/Microsoft/Edge/Application/msedge.exe", "--app=http://localhost:4300").Run()
	if err == nil {
		return true
	}
	err = exec.Command("C:/Program Files (x86)/Microsoft/Edge/Application/msedge.exe", "--app=http://localhost:4300").Run()
	if err == nil {
		return true
	}
	err = exec.Command("C:/Program Files/Google/Chrome/Application/chrome.exe", "--app=http://localhost:4300").Run()
	if err == nil {
		return true
	}
	err = exec.Command("C:/Program Files (x86)/Google/Chrome/Application/chrome.exe", "--app=http://localhost:4300").Run()
	if err == nil {
		return true
	}
	err = browser.OpenURL("http://localhost:4300")
	return err == nil
}

func nix() bool {
	err := exec.Command("google-chrome", "--app=http://localhost:4300").Run()
	if err == nil {
		return true
	}
	err = exec.Command("google-chrome-stable", "--app=http://localhost:4300").Run()
	if err == nil {
		return true
	}
	err = exec.Command("chromium", "--app=http://localhost:4300").Run()
	if err == nil {
		return true
	}
	err = exec.Command("chromium-browser", "--app=http://localhost:4300").Run()
	if err == nil {
		return true
	}
	err = exec.Command("/usr/bin/google-chrome", "--app=http://localhost:4300").Run()
	if err == nil {
		return true
	}
	err = exec.Command("/usr/bin/google-chrome-stable", "--app=http://localhost:4300").Run()
	if err == nil {
		return true
	}
	err = exec.Command("/usr/bin/chromium", "--app=http://localhost:4300").Run()
	if err == nil {
		return true
	}
	err = exec.Command("/usr/bin/chromium-browser", "--app=http://localhost:4300").Run()
	if err == nil {
		return true
	}
	err = exec.Command("microsoft-edge", "--app=http://localhost:4300").Run()
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
	if err == nil {
		return true
	}
	// Firefox can no longer be launched without an address bar as the ssb parameter (site
	// specific browser) was removed in 2021.
	err = exec.Command("firefox", "--new-window", "http://localhost:4300").Run()
	if err == nil {
		return true
	}
	err = exec.Command("/usr/bin/firefox", "--new-window", "http://localhost:4300").Run()
	if err == nil {
		return true
	}
	err = browser.OpenURL("http://localhost:4300")
	return err == nil
}
