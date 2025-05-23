package payment

import (
	"context"
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/google/uuid"

	"github.com/genefriendway/onchain-handler/constants"
	ucasetypes "github.com/genefriendway/onchain-handler/internal/domain/ucases/types"
	"github.com/genefriendway/onchain-handler/pkg/crypto"
	"github.com/genefriendway/onchain-handler/pkg/logger"
)

// Function to generate wallets and insert them into the database if none exist
func InitPaymentWallets(
	ctx context.Context,
	mnemonic, passphrase, salt string,
	totalWallets uint,
	walletUCase ucasetypes.PaymentWalletUCase,
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

// Function to get the receiving wallet
func GetReceivingWallet(mnemonic, passphrase, salt string) (*accounts.Account, *ecdsa.PrivateKey, error) {
	// Generate receiving wallet
	account, privateKey, err := crypto.GenerateAccount(mnemonic, passphrase, salt, constants.ReceivingWallet, 0)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate master wallet: %w", err)
	}

	return account, privateKey, nil
}

func GenerateTempAddress() string {
	uuidPart := uuid.New().String()
	hash := sha256.Sum256([]byte(uuidPart))           // Hash UUID for uniqueness
	return "temp-" + hex.EncodeToString(hash[:])[:37] // Ensure length <= 42
}
