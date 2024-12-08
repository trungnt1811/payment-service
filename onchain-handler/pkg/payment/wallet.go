package payment

import (
	"context"
	"fmt"

	"github.com/genefriendway/onchain-handler/internal/interfaces"
	"github.com/genefriendway/onchain-handler/pkg/logger"
)

// Function to generate wallets and insert them into the database if none exist
func InitPaymentWallets(
	ctx context.Context,
	mnemonic, passphrase, salt string,
	totalWallets uint,
	walletUCase interfaces.PaymentWalletUCase,
) error {
	// Check if wallets already exist
	isExist, err := walletUCase.IsRowExist(ctx)
	if err != nil {
		return fmt.Errorf("failed to check existing wallets: %w", err)
	}

	if isExist {
		logger.GetLogger().Info("Payment wallets already exist")
		return nil
	}

	// Insert wallets if none exist
	for i := 0; i < int(totalWallets); i++ {
		inUse := false
		err := walletUCase.CreateAndGenerateWallet(ctx, mnemonic, passphrase, salt, inUse)
		if err != nil {
			return fmt.Errorf("failed to initialize wallet %d: %w", i+1, err)
		}
	}

	logger.GetLogger().Infof("Successfully created %d payment wallets", totalWallets)
	return nil
}
