package payment_order

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/genefriendway/onchain-handler/constants"
	"github.com/genefriendway/onchain-handler/infra/caching"
	"github.com/genefriendway/onchain-handler/internal/dto"
	"github.com/genefriendway/onchain-handler/internal/interfaces"
	"github.com/genefriendway/onchain-handler/internal/model"
)

type paymentOrderUCase struct {
	db                      *gorm.DB                           // Gorm DB instance for transaction handling
	paymentOrderRepository  interfaces.PaymentOrderRepository  // Repository to interact with payment orders
	paymentWalletRepository interfaces.PaymentWalletRepository // Repository to interact with payment wallets
	blockStateRepo          interfaces.BlockStateRepository    // Repository to fetch the latest block state
	cacheRepo               caching.CacheRepository            // Cache repository to retrieve and store block information
}

// NewPaymentOrderUCase constructs a new paymentOrderUCase with the provided dependencies.
func NewPaymentOrderUCase(
	db *gorm.DB, // Add DB to the constructor
	paymentOrderRepository interfaces.PaymentOrderRepository,
	paymentWalletRepository interfaces.PaymentWalletRepository,
	blockStateRepo interfaces.BlockStateRepository,
	cacheRepo caching.CacheRepository,
) interfaces.PaymentOrderUCase {
	return &paymentOrderUCase{
		db:                      db,
		paymentOrderRepository:  paymentOrderRepository,
		paymentWalletRepository: paymentWalletRepository,
		blockStateRepo:          blockStateRepo,
		cacheRepo:               cacheRepo,
	}
}

// CreatePaymentOrders creates new payment orders from the given payloads.
// It first attempts to retrieve the latest block from cache. If not found in cache, it fetches it from the database.
// Each payload is transformed into a PaymentOrder model, setting the block height to the latest block.
func (u *paymentOrderUCase) CreatePaymentOrders(
	ctx context.Context,
	payloads []dto.PaymentOrderPayloadDTO,
	expiredOrderTime time.Duration,
) ([]dto.PaymentOrderDTO, error) {
	cacheKey := &caching.Keyer{Raw: constants.LatestBlockCacheKey} // Cache key for the latest block

	// Try to retrieve the latest block from cache
	var latestBlock uint64
	err := u.cacheRepo.RetrieveItem(cacheKey, &latestBlock)
	if err != nil {
		// If cache is empty, load from the database
		latestBlock, err = u.blockStateRepo.GetLatestBlock(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to retrieve latest block from database: %w", err)
		}
	}

	var response []dto.PaymentOrderDTO

	// Begin a new transaction
	err = u.db.Transaction(func(tx *gorm.DB) error {
		var orders []model.PaymentOrder

		// Convert each payload to a PaymentOrder model, associating it with the latest block
		for _, payload := range payloads {
			// Try to claim an available wallet
			assignWallet, err := u.paymentWalletRepository.ClaimFirstAvailableWallet(ctx)
			if err != nil {
				return fmt.Errorf("failed to claim available wallet: %w", err)
			}

			// Create the PaymentOrder model for the current payload
			order := model.PaymentOrder{
				Amount:      payload.Amount,
				WalletID:    assignWallet.ID, // Associate the order with the claimed wallet
				Transferred: "0",             // Default transferred status
				UserID:      payload.UserID,
				BlockHeight: latestBlock,       // Use the latest block height
				Status:      constants.Pending, // Set order status to pending
				ExpiredTime: time.Now().UTC().Add(expiredOrderTime),
			}
			orders = append(orders, order)

			orderDTO := order.ToDto()
			orderDTO.PaymentAddress = assignWallet.Address
			response = append(response, orderDTO)
		}

		// Save the created payment orders to the database within the transaction
		orders, err = u.paymentOrderRepository.CreatePaymentOrders(ctx, orders)
		if err != nil {
			return fmt.Errorf("failed to create payment orders: %w", err)
		}

		// Mappping order id
		for index, order := range orders {
			response[index].ID = order.ID
		}

		return nil
	})
	// Check if there was any error within the transaction
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (u *paymentOrderUCase) UpdateExpiredOrdersToFailed(ctx context.Context) error {
	return u.paymentOrderRepository.UpdateExpiredOrdersToFailed(ctx)
}

func (u *paymentOrderUCase) GetExpiredPaymentOrders(ctx context.Context) ([]model.PaymentOrder, error) {
	return u.paymentOrderRepository.GetExpiredPaymentOrders(ctx)
}

func (u *paymentOrderUCase) UpdatePaymentOrder(
	ctx context.Context,
	orderID uint64,
	status, transferredAmount string,
	walletStatus bool,
	blockHeight uint64,
) error {
	return u.paymentOrderRepository.UpdatePaymentOrder(ctx, orderID, status, transferredAmount, walletStatus, blockHeight)
}

func (u *paymentOrderUCase) BatchUpdateOrderStatuses(ctx context.Context, orders []dto.PaymentOrderDTO) error {
	var orderIDs []uint64
	var newStatuses []string
	// Parse orders to extract orderIDs and newStatuses
	for _, order := range orders {
		orderIDs = append(orderIDs, order.ID)
		newStatuses = append(newStatuses, order.Status)
	}

	return u.paymentOrderRepository.BatchUpdateOrderStatuses(ctx, orderIDs, newStatuses)
}

func (u *paymentOrderUCase) GetPendingPaymentOrders(ctx context.Context, size int) ([]dto.PaymentOrderDTO, error) {
	orders, err := u.paymentOrderRepository.GetPendingPaymentOrders(ctx, size, 0)
	if err != nil {
		return nil, err
	}
	var orderDtos []dto.PaymentOrderDTO
	for _, order := range orders {
		orderDtos = append(orderDtos, order.ToDto())
	}
	return orderDtos, nil
}
