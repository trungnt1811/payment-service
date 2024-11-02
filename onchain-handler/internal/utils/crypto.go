package utils

import (
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"fmt"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/tyler-smith/go-bip32"
	"github.com/tyler-smith/go-bip39"

	"github.com/genefriendway/onchain-handler/constants"
)

// SignMessage signs a message using a private key and returns the signature
func SignMessage(privateKey *ecdsa.PrivateKey, message []byte) (singnature []byte, err error) {
	if privateKey == nil {
		return nil, fmt.Errorf("private key is nil")
	}
	if message == nil {
		return nil, fmt.Errorf("message is nil")
	}

	// Hash the message using Keccak256
	hash := crypto.Keccak256Hash(message)

	// Recover from panics and return an error
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic: %v", r)
		}
	}()

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

func GenerateAccount(mnemonic, passphrase, salt string, walletType constants.WalletType, id uint64) (*accounts.Account, *ecdsa.PrivateKey, error) {
	// Generate the seed from the mnemonic and passphrase
	seed := bip39.NewSeed(mnemonic, passphrase)

	// Create the master key from the seed
	masterKey, err := bip32.NewMasterKey(seed)
	if err != nil {
		return nil, nil, err
	}

	// Convert walletType to a unique integer for use in the HD Path
	walletTypeHash := HashToUint32(string(walletType) + fmt.Sprint(id))

	// Define the HD Path for Ethereum address (e.g., m/44'/60'/id'/walletTypeHash/salt)
	path := []uint32{
		44 + bip32.FirstHardenedChild,         // BIP44 purpose field
		60 + bip32.FirstHardenedChild,         // Ethereum coin type
		uint32(id) + bip32.FirstHardenedChild, // User-specific field
		walletTypeHash,                        // Unique integer based on wallet type and id
		HashToUint32(salt),                    // Hash of salt for additional security
	}

	// Derive a private key along the specified HD Path
	key := masterKey
	for _, index := range path {
		key, err = key.NewChildKey(index)
		if err != nil {
			return nil, nil, err
		}
	}

	// Generate an Ethereum account from the derived private key
	privateKey, err := crypto.ToECDSA(key.Key)
	if err != nil {
		return nil, nil, err
	}

	account := accounts.Account{
		Address: crypto.PubkeyToAddress(privateKey.PublicKey),
	}

	return &account, privateKey, nil
}
