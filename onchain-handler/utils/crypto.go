package utils

import (
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"fmt"

	"github.com/ethereum/go-ethereum/crypto"
)

// SignMessage signs a message using a private key and returns the signature
func SignMessage(privateKey *ecdsa.PrivateKey, message []byte) ([]byte, error) {
	// Hash the message using Keccak256
	hash := crypto.Keccak256Hash(message)

	// Sign the hash with the private key
	signature, err := crypto.Sign(hash.Bytes(), privateKey)
	if err != nil {
		return nil, err
	}

	return signature, nil
}

// VerifySignature verifies the signature against an Ethereum address and message
func VerifySignature(address, signature string, message []byte) (bool, error) {
	// Hash the message
	hash := crypto.Keccak256Hash(message)

	// Decode the base64-encoded signature string
	signatureBytes, err := decodeSignature(signature)
	if err != nil {
		return false, err
	}

	// Recover the public key from the signature
	pubKey, err := crypto.Ecrecover(hash.Bytes(), signatureBytes)
	if err != nil {
		return false, err
	}

	// Convert public key to an Ethereum address
	recoveredPubKey, err := crypto.UnmarshalPubkey(pubKey)
	if err != nil {
		return false, err
	}
	recoveredAddress := crypto.PubkeyToAddress(*recoveredPubKey).Hex()

	// Compare recovered address to expected address
	return recoveredAddress == address, nil
}

func decodeSignature(signature string) ([]byte, error) {
	// Decode the base64-encoded signature string
	signatureBytes, err := base64.StdEncoding.DecodeString(signature)
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64 signature: %w", err)
	}
	return signatureBytes, nil
}

func HashToUint32(input string) uint32 {
	hash := sha256.Sum256([]byte(input))
	return binary.BigEndian.Uint32(hash[:4])
}

// PrivateKeyFromHex converts a private key string in hex format to an ECDSA private key
func PrivateKeyFromHex(privateKeyHex string) (*ecdsa.PrivateKey, error) {
	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		return nil, fmt.Errorf("failed to convert private key from hex: %w", err)
	}
	return privateKey, nil
}
