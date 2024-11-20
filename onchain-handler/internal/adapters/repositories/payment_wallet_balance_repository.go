package repositories

import (
	"context"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/genefriendway/onchain-handler/internal/interfaces"
)

type paymentWalletBalanceRepository struct {
	db *gorm.DB
}

func NewPaymentWalletBalanceRepository(db *gorm.DB) interfaces.PaymentWalletBalanceRepository {
	return &paymentWalletBalanceRepository{
		db: db,
	}
}

// UpsertPaymentWalletBalances updates or inserts the balances of multiple wallets by their WalletIDs.
func (r *paymentWalletBalanceRepository) UpsertPaymentWalletBalances(
	ctx context.Context,
	walletIDs []uint64,
	newBalances []string,
	network string,
	symbol string,
) error {
	// Check if the lengths of walletIDs and newBalances match
	if len(walletIDs) != len(newBalances) {
		return fmt.Errorf("the number of wallet IDs and balances must be the same")
	}

	// Build the upsert query with placeholders for batch insertion
	values := []string{}
	args := []interface{}{}
	for i, walletID := range walletIDs {
		values = append(values, fmt.Sprintf("($%d, $%d, $%d, $%d, $%d)", i*5+1, i*5+2, i*5+3, i*5+4, i*5+5))
		args = append(args, walletID, network, symbol, newBalances[i], time.Now())
	}

	query := `
		INSERT INTO payment_wallet_balance (wallet_id, network, symbol, balance, updated_at)
		VALUES ` + strings.Join(values, ",") + `
		ON CONFLICT (wallet_id, network, symbol)
		DO UPDATE SET 
			balance = EXCLUDED.balance;
	`

	// Execute the query
	if err := r.db.WithContext(ctx).Exec(query, args...).Error; err != nil {
		return fmt.Errorf("failed to upsert wallet balances: %w", err)
	}

	return nil
}
