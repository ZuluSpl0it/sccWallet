package build

import (
	"os"
	"path/filepath"
	"runtime"
)

// ScPrimeWebWalletDir returns the ScPrime web wallet's data directory either from·
// the environment variable or the default.
func ScPrimeWebWalletDir() string {
	dataDir := os.Getenv(EnvvarMetaDataDir)
	if dataDir == "" {
		return defaultScPrimeWebWalletDir()
	}
	return dataDir
}

// defaultScPrimeWebWalletDir returns the default data directory of scp-webwallet.
// The values for supported operating systems are:
//
// Linux:   $HOME/.scprime-webwallet
// MacOS:   $HOME/Library/Application Support/ScPrime-WebWallet
// Windows: %LOCALAPPDATA%\ScPrime-WebWallet
func defaultScPrimeWebWalletDir() string {
	switch runtime.GOOS {
	case "windows":
		return filepath.Join(os.Getenv("LOCALAPPDATA"), "ScPrime-WebWallet")
	case "darwin":
		return filepath.Join(os.Getenv("HOME"), "Library", "Application Support", "ScPrime-WebWallet")
	default:
		return filepath.Join(os.Getenv("HOME"), ".scprime-webwallet")
	}
}
