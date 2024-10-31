package utils

import (
	"crypto/ecdsa"
	"fmt"

	"github.com/ethereum/go-ethereum/crypto"
)

// PrivateKeyFromHex converts a private key string in hex format to an ECDSA private key
func PrivateKeyFromHex(privateKeyHex string) (*ecdsa.PrivateKey, error) {
	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		return nil, fmt.Errorf("failed to convert private key from hex: %w", err)
	}
	return privateKey, nil
}
