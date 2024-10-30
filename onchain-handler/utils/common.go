package utils

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/tyler-smith/go-bip32"
	"github.com/tyler-smith/go-bip39"

	"github.com/genefriendway/onchain-handler/conf"
	"github.com/genefriendway/onchain-handler/constants"
	"github.com/genefriendway/onchain-handler/internal/interfaces"
	"github.com/genefriendway/onchain-handler/internal/model"
)

func CalculatePaymentCovering(amount *big.Int, paymentCoveringFactor float64) *big.Int {
	// Convert the covering factor to a multiplier as a float64.
	coveringFactorMultiplier := 1 - (paymentCoveringFactor / 100)

	amountFloat := new(big.Float).SetInt(amount)
	coveringFactorFloat := big.NewFloat(coveringFactorMultiplier)

	// Perform the multiplication (big.Float * big.Float)
	minimumAcceptedAmountFloat := new(big.Float).Mul(amountFloat, coveringFactorFloat)

	// Convert the result back to a big.Int (this rounds the float result)
	minimumAcceptedAmount := new(big.Int)
	minimumAcceptedAmountFloat.Int(minimumAcceptedAmount)
	return minimumAcceptedAmount
}

// Function to generate wallets and insert them into the database if none exist
func InitPaymentWallets(
	ctx context.Context,
	config *conf.Configuration,
	walletRepo interfaces.PaymentWalletRepository,
) error {
	// Check if wallets already exist in the database
	isExist, err := walletRepo.IsRowExist(ctx)
	if err != nil {
		return err
	}

	// Insert wallets into the database if none exist
	if !isExist {
		var wallets []model.PaymentWallet
		initWalletCount := config.PaymentGateway.InitWalletCount
		for index := 1; index <= int(initWalletCount); index++ {
			account, _, err := GenerateAccount(
				config.Wallet.Mnemonic,
				config.Wallet.Passphrase,
				config.Wallet.Salt,
				constants.PaymentWallet,
				uint64(index),
			)
			if err != nil {
				return err
			}
			wallet := model.PaymentWallet{
				ID:      uint64(index),
				Address: account.Address.Hex(),
				InUse:   false, // New wallets are not in use by default
			}
			wallets = append(wallets, wallet)
		}

		err := walletRepo.CreatePaymentWallets(ctx, wallets)
		if err != nil {
			return err
		}
		log.Println("Successfully created payment wallets")
	}

	return nil
}

func GenerateAccount(mnemonic, passphrase, salt, walletType string, id uint64) (*accounts.Account, *ecdsa.PrivateKey, error) {
	// Generate the seed from the mnemonic and passphrase
	seed := bip39.NewSeed(mnemonic, passphrase)

	// Create the master key from the seed
	masterKey, err := bip32.NewMasterKey(seed)
	if err != nil {
		return nil, nil, err
	}

	// Convert walletType to a unique integer for use in the HD Path
	walletTypeHash := HashToUint32(walletType + fmt.Sprint(id))

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

	// privateKeyBytes := crypto.FromECDSA(privateKey)
	// privateKeyStr := hex.EncodeToString(privateKeyBytes)
	// println(privateKeyStr)
	// println(account.Address.Hex())

	return &account, privateKey, nil
}
