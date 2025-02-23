package providers

import (
	"context"
	"sync"

	"github.com/genefriendway/onchain-handler/internal/adapters/orderset"
	settypes "github.com/genefriendway/onchain-handler/internal/adapters/orderset/types"
	"github.com/genefriendway/onchain-handler/internal/delivery/dto"
	"github.com/genefriendway/onchain-handler/pkg/logger"
)

var (
	paymentOrderSetOnce sync.Once
	paymentOrderSet     settypes.Set[dto.PaymentOrderDTO]
)

// ProvidePaymentOrderSet initializes and provides a singleton instance of PaymentOrderSet.
func ProvidePaymentOrderSet(ctx context.Context) settypes.Set[dto.PaymentOrderDTO] {
	paymentOrderSetOnce.Do(func() {
		keyFunc := func(order dto.PaymentOrderDTO) string {
			return order.PaymentAddress + "_" + order.Symbol
		}

		// Initialize the order set
		var err error
		paymentOrderSet, err = orderset.NewSet(ctx, keyFunc)
		if err != nil {
			logger.GetLogger().Fatalf("Create payment order set error: %v", err)
		}
	})
	return paymentOrderSet
}
