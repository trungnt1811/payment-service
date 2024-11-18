package crypto

import (
	"crypto/ecdsa"
	"encoding/base64"
	"testing"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/require"

	"github.com/genefriendway/onchain-handler/constants"
)

func TestSignMessage(t *testing.T) {
	t.Run("SignMessage", func(t *testing.T) {
		privateKey, err := crypto.GenerateKey()
		require.NoError(t, err)

		message := []byte("test message")
		signature, err := SignMessage(privateKey, message)
		require.NoError(t, err)
		require.NotNil(t, signature)
	})

	t.Run("NilPrivateKey", func(t *testing.T) {
		message := []byte("test message")
		signature, err := SignMessage(nil, message)
		require.Error(t, err)
		require.Nil(t, signature)
	})

	t.Run("NilMessage", func(t *testing.T) {
		privateKey, err := crypto.GenerateKey()
		require.NoError(t, err)

		signature, err := SignMessage(privateKey, nil)
		require.Error(t, err)
		require.Nil(t, signature)
	})

	t.Run("InvalidPrivateKey", func(t *testing.T) {
		invalidPrivateKey := &ecdsa.PrivateKey{}
		message := []byte("test message")
		signature, err := SignMessage(invalidPrivateKey, message)
		require.Error(t, err)
		require.Nil(t, signature)
	})
}

func TestVerifySignature(t *testing.T) {
	t.Run("VerifySignature", func(t *testing.T) {
		privateKey, err := crypto.GenerateKey()
		require.NoError(t, err)

		message := []byte("test message")
		signature, err := SignMessage(privateKey, message)
		require.NoError(t, err)

		address := crypto.PubkeyToAddress(privateKey.PublicKey).Hex()
		signatureBase64 := base64.StdEncoding.EncodeToString(signature)

		valid, err := VerifySignature(address, signatureBase64, message)
		require.NoError(t, err)
		require.True(t, valid)
	})

	t.Run("InvalidSignature", func(t *testing.T) {
		privateKey, err := crypto.GenerateKey()
		require.NoError(t, err)

		message := []byte("test message")
		_, err = SignMessage(privateKey, message)
		require.NoError(t, err)

		address := crypto.PubkeyToAddress(privateKey.PublicKey).Hex()
		invalidSignature := base64.StdEncoding.EncodeToString([]byte("invalid signature"))

		valid, err := VerifySignature(address, invalidSignature, message)
		require.Error(t, err)
		require.False(t, valid)
	})

	t.Run("InvalidAddress", func(t *testing.T) {
		privateKey, err := crypto.GenerateKey()
		require.NoError(t, err)

		message := []byte("test message")
		signature, err := SignMessage(privateKey, message)
		require.NoError(t, err)

		address := "invalid address"
		signatureBase64 := base64.StdEncoding.EncodeToString(signature)

		valid, err := VerifySignature(address, signatureBase64, message)
		require.NoError(t, err)
		require.False(t, valid)
	})

	t.Run("InvalidSignatureLength", func(t *testing.T) {
		privateKey, err := crypto.GenerateKey()
		require.NoError(t, err)

		message := []byte("test message")
		signature, err := SignMessage(privateKey, message)
		require.NoError(t, err)

		address := crypto.PubkeyToAddress(privateKey.PublicKey).Hex()
		signatureBase64 := base64.StdEncoding.EncodeToString(signature[:len(signature)-1])

		valid, err := VerifySignature(address, signatureBase64, message)
		require.Error(t, err)
		require.False(t, valid)
	})

	t.Run("InvalidDecodedSignature", func(t *testing.T) {
		privateKey, err := crypto.GenerateKey()
		require.NoError(t, err)

		address := crypto.PubkeyToAddress(privateKey.PublicKey).Hex()
		message := []byte("test message")
		invalidSignature := "invalid signature"

		valid, err := VerifySignature(address, invalidSignature, message)
		require.Error(t, err)
		require.False(t, valid)
	})
}

func TestHashToUint32(t *testing.T) {
	t.Run("HashToUint32", func(t *testing.T) {
		input := "test input"
		expected := uint32(0x9dfe6f15) // Precomputed hash value
		result := HashToUint32(input)
		require.Equal(t, expected, result)
	})
}

func TestGenerateAccount(t *testing.T) {
	t.Run("TestGenerateAccount", func(t *testing.T) {
		mnemonic := "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"
		passphrase := ""
		salt := "test-salt"
		walletType := constants.WalletType("test-wallet")
		id := uint64(1)

		account, privateKey, err := GenerateAccount(mnemonic, passphrase, salt, walletType, id)
		require.NoError(t, err)
		require.NotNil(t, account)
		require.NotNil(t, privateKey)
		require.Equal(t, account.Address, crypto.PubkeyToAddress(privateKey.PublicKey))
	})
}

func TestPubkeyToAddress(t *testing.T) {
	t.Run("TestPubkeyToAddress", func(t *testing.T) {
		privateKey, err := crypto.GenerateKey()
		require.NoError(t, err)

		address := crypto.PubkeyToAddress(privateKey.PublicKey)
		require.NotNil(t, address)
	})
}

func TestPrivateKeyFromHex(t *testing.T) {
	t.Run("Valid hex string", func(t *testing.T) {
		hexKey := "4c0883a69102937d6231471b5dbb6204fe512961708279f1d7e1b8d3e5e8a6e4"
		privateKey, err := PrivateKeyFromHex(hexKey)
		require.NoError(t, err)
		require.NotNil(t, privateKey)
	})

	t.Run("Invalid hex string", func(t *testing.T) {
		hexKey := "invalid-hex-string"
		privateKey, err := PrivateKeyFromHex(hexKey)
		require.Error(t, err)
		require.Nil(t, privateKey)
	})

	t.Run("Empty hex string", func(t *testing.T) {
		hexKey := ""
		privateKey, err := PrivateKeyFromHex(hexKey)
		require.Error(t, err)
		require.Nil(t, privateKey)
	})
}
