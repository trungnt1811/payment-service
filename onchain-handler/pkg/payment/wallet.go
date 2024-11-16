package payment

import (
	"context"

	"github.com/genefriendway/onchain-handler/conf"
	"github.com/genefriendway/onchain-handler/constants"
	"github.com/genefriendway/onchain-handler/internal/dto"
	"github.com/genefriendway/onchain-handler/internal/interfaces"
	"github.com/genefriendway/onchain-handler/pkg/crypto"
	"github.com/genefriendway/onchain-handler/pkg/logger"
)

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
			account, _, err := crypto.GenerateAccount(
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
		logger.GetLogger().Info("Successfully created payment wallets")
	}

	return nil
}
