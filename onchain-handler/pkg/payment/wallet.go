package payment

import (
	"context"
	"crypto/ecdsa"

	"github.com/ethereum/go-ethereum/accounts"

	"github.com/genefriendway/onchain-handler/constants"
	"github.com/genefriendway/onchain-handler/internal/dto"
	"github.com/genefriendway/onchain-handler/internal/interfaces"
	"github.com/genefriendway/onchain-handler/pkg/crypto"
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
		return err
	}

	// Insert wallets if none exist
	if !isExist {
		var wallets []dto.PaymentWalletPayloadDTO
		for index := 1; index <= int(totalWallets); index++ {
			account, _, err := crypto.GenerateAccount(
				mnemonic,
				passphrase,
				salt,
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

func RetriveReceivingWallet(mnemonic, passphrase, salt string) (*accounts.Account, *ecdsa.PrivateKey, error) {
	return crypto.GenerateAccount(mnemonic, passphrase, salt, constants.ReceivingWallet, 0)
}
