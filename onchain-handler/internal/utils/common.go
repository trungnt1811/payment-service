package utils

import (
	"context"
	"math/big"

	"github.com/genefriendway/onchain-handler/conf"
	"github.com/genefriendway/onchain-handler/constants"
	"github.com/genefriendway/onchain-handler/internal/dto"
	"github.com/genefriendway/onchain-handler/internal/interfaces"
	"github.com/genefriendway/onchain-handler/log"
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
	walletUCase interfaces.PaymentWalletUCase,
) error {
	// Check if wallets already exist
	isExist, err := walletUCase.IsRowExist(ctx)
	if err != nil {
		return err
	}

	// Insert wallets if none exist
	if !isExist {
		var wallets []dto.PaymentWalletPayloadDTO
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
			wallet := dto.PaymentWalletPayloadDTO{
				ID:      uint64(index),
				Address: account.Address.Hex(),
				InUse:   false, // New wallets are not in use by default
			}
			wallets = append(wallets, wallet)
		}

		err := walletUCase.CreatePaymentWallets(ctx, wallets)
		if err != nil {
			return err
		}
		log.LG.Info("Successfully created payment wallets")
	}

	return nil
}
