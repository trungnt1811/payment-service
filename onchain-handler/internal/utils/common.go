package utils

import (
	"context"
	"log"

	"github.com/genefriendway/onchain-handler/conf"
	"github.com/genefriendway/onchain-handler/internal/interfaces"
	"github.com/genefriendway/onchain-handler/internal/model"
)

// Function to generate wallets and insert them into the database if none exist
func InitPaymentWallets(
	ctx context.Context,
	config *conf.Configuration,
	walletRepo interfaces.PaymentWalletRepository,
) error {
	wallets, err := GenerateWallets(config, config.PaymentGateway.InitWalletCount)
	if err != nil {
		return err
	}

	// Check if wallets already exist in the database
	isExist, err := walletRepo.IsRowExist(ctx)
	if err != nil {
		return err
	}

	// Insert wallets into the database if none exist
	if !isExist {
		err := walletRepo.CreatePaymentWallets(ctx, wallets)
		if err != nil {
			return err
		}
		log.Println("Successfully created payment wallets")
	}

	return nil
}

// TODO: consider using an HDPath (Hierarchical Deterministic Path) to create multiple wallets from a single mnemonic.
// GenerateWallets creates a specified number of payment wallets and encrypts the private keys
func GenerateWallets(config *conf.Configuration, count uint) ([]model.PaymentWallet, error) {
	encryptionKey := config.GetEncryptionKey()
	var wallets []model.PaymentWallet

	for i := 0; i < int(count); i++ {
		privateKeyHex, address, err := GenerateKeyPair()
		if err != nil {
			return nil, err
		}

		// Encrypt the private key before storing it
		encryptedPrivateKey, err := Encrypt(privateKeyHex, encryptionKey)
		if err != nil {
			return nil, err
		}

		wallet := model.PaymentWallet{
			Address:    address,
			PrivateKey: encryptedPrivateKey, // Store encrypted private key
			InUse:      false,               // New wallets are not in use by default
		}
		wallets = append(wallets, wallet)
	}

	return wallets, nil
}
