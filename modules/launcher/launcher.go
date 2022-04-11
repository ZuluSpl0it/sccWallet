package launcher

import (
	"runtime"

	"gitlab.com/scpcorp/webwallet/modules/launcher/darwin"
	"gitlab.com/scpcorp/webwallet/modules/launcher/nix"
	"gitlab.com/scpcorp/webwallet/modules/launcher/windows"
)

// Launch will attempt to launch the application in the supplied browser. If that fails the
// launcher will attempt to launch the application in a series of fallback browsers. Browsers
// that are based on Chromium (such as Google Chrome and Microsoft Edge) are most desirable
// because they can be launched in app mode (which means that there is no address bar). This
// allows the GUI head feel most like a native application.
func Launch(browser string) bool {
	switch runtime.GOOS {
	case "darwin":
		return darwin.Launch(browser)
	case "windows":
		return windows.Launch(browser)
	}
	return nix.Launch(browser)
}
