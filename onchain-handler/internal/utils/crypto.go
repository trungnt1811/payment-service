package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/ethereum/go-ethereum/crypto"
)

// Encrypt text using AES-GCM encryption
func Encrypt(data, encryptionKeyBase64 string) (string, error) {
	if encryptionKeyBase64 == "" {
		return "", errors.New("encryption key is not set")
	}

	// Decode the base64 encoded key to get the actual byte array
	encryptionKey, err := base64.StdEncoding.DecodeString(encryptionKeyBase64)
	if err != nil {
		return "", err
	}

	// Ensure the key is 32 bytes (for AES-256)
	if len(encryptionKey) != 32 {
		return "", errors.New("invalid key length: key must be 32 bytes (AES-256)")
	}

	// Create AES cipher block
	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return "", err
	}

	// Use GCM mode for AES encryption
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	// Generate a nonce (random value)
	nonce := make([]byte, aesGCM.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	// Encrypt the plaintext
	cipherText := aesGCM.Seal(nonce, nonce, []byte(data), nil)

	// Return the base64-encoded ciphertext
	return base64.StdEncoding.EncodeToString(cipherText), nil
}

// Decrypt encrypted text using AES-GCM
func Decrypt(encryptedText string) (string, error) {
	// Get the base64 encoded encryption key from the environment variable
	encryptionKeyBase64 := os.Getenv("ENCRYPTION_KEY")
	if encryptionKeyBase64 == "" {
		return "", errors.New("encryption key is not set")
	}

	// Decode the base64 encoded key
	encryptionKey, err := base64.StdEncoding.DecodeString(encryptionKeyBase64)
	if err != nil {
		return "", err
	}

	// Ensure the key is 32 bytes (for AES-256)
	if len(encryptionKey) != 32 {
		return "", errors.New("invalid key length: key must be 32 bytes (AES-256)")
	}

	// Decode the base64-encoded ciphertext
	cipherText, err := base64.StdEncoding.DecodeString(encryptedText)
	if err != nil {
		return "", err
	}

	// Create AES cipher block
	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return "", err
	}

	// Use GCM mode for decryption
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := aesGCM.NonceSize()
	if len(cipherText) < nonceSize {
		return "", errors.New("ciphertext too short")
	}

	// Extract the nonce and actual ciphertext
	nonce, cipherText := cipherText[:nonceSize], cipherText[nonceSize:]

	// Decrypt the ciphertext
	plainText, err := aesGCM.Open(nil, nonce, cipherText, nil)
	if err != nil {
		return "", err
	}

	return string(plainText), nil
}

// GenerateKeyPair generates a new private key and associated address.
func GenerateKeyPair() (privateKeyStr, address string, err error) {
	// Generate a new private key
	privateKey, err := ecdsa.GenerateKey(crypto.S256(), rand.Reader)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate private key: %w", err)
	}

	// Convert the private key to its string form (without "0x" prefix)
	privateKeyBytes := crypto.FromECDSA(privateKey)
	privateKeyStr = hex.EncodeToString(privateKeyBytes) // Encodes private key as string without "0x"

	// Derive the public key from the private key
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return "", "", fmt.Errorf("failed to cast public key to ECDSA")
	}

	// Derive the address from the public key
	address = crypto.PubkeyToAddress(*publicKeyECDSA).Hex()

	return privateKeyStr, address, nil
}
