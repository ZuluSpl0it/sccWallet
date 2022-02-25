package main

import (
	"fmt"
	"os"

	"golang.org/x/crypto/ssh/terminal"

	"gitlab.com/scpcorp/ScPrime/crypto"
	"gitlab.com/scpcorp/ScPrime/modules"
	"gitlab.com/scpcorp/ScPrime/node"

	mnemonics "gitlab.com/NebulousLabs/entropy-mnemonics"
)

func unlock(node *node.Node) (string, error) {
	// Is wallet encrypted?
	encrypted, err := node.Wallet.Encrypted()
	if err != nil {
		return "Unable to determine if wallet is encrypted", err
	}
	fmt.Println("Type Wallet Password:")
	password, _ := terminal.ReadPassword(int(os.Stdin.Fd()))
	if !encrypted {
		fmt.Println("Initializing wallet...")
		encryptionKey := crypto.NewWalletKey(crypto.HashObject(string(password)))
		_, err = node.Wallet.Encrypt(encryptionKey)
		if err != nil {
			return "Unable to initialize new wallet seed", err
		}
	}
	potentialKeys, _ := encryptionKeys(string(password))
	for _, key := range potentialKeys {
		unlocked, unlockingErr := node.Wallet.Unlocked()
		if unlockingErr != nil {
			return "Unable to initialize new wallet seed", unlockingErr
		}
		if !unlocked {
			unlockingErr = node.Wallet.Unlock(key)
			if unlockingErr != nil {
				return "Unable to initialize new wallet seed", unlockingErr
			}
		}
	}
	return "Wallet is unlocked...", nil
}

// encryptionKeys enumerates the possible encryption keys that can be derived
// from an input string.
// copied from gitlab.com/scpcorp/ScPrime/node/api/wallet.go
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
