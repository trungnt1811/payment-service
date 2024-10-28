package payment_order

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/genefriendway/onchain-handler/conf"
	"github.com/genefriendway/onchain-handler/constants"
	"github.com/genefriendway/onchain-handler/infra/caching"
	"github.com/genefriendway/onchain-handler/internal/dto"
	"github.com/genefriendway/onchain-handler/internal/interfaces"
	"github.com/genefriendway/onchain-handler/internal/model"
	internalutils "github.com/genefriendway/onchain-handler/internal/utils"
	"github.com/genefriendway/onchain-handler/utils"
)

type paymentOrderUCase struct {
	db                      *gorm.DB                           // Gorm DB instance for transaction handling
	paymentOrderRepository  interfaces.PaymentOrderRepository  // Repository to interact with payment orders
	paymentWalletRepository interfaces.PaymentWalletRepository // Repository to interact with payment wallets
	blockStateRepo          interfaces.BlockStateRepository    // Repository to fetch the latest block state
	cacheRepo               caching.CacheRepository            // Cache repository to retrieve and store block information
	config                  *conf.Configuration
}

// NewPaymentOrderUCase constructs a new paymentOrderUCase with the provided dependencies.
func NewPaymentOrderUCase(
	db *gorm.DB, // Add DB to the constructor
	paymentOrderRepository interfaces.PaymentOrderRepository,
	paymentWalletRepository interfaces.PaymentWalletRepository,
	blockStateRepo interfaces.BlockStateRepository,
	cacheRepo caching.CacheRepository,
	config *conf.Configuration,
) interfaces.PaymentOrderUCase {
	return &paymentOrderUCase{
		db:                      db,
		paymentOrderRepository:  paymentOrderRepository,
		paymentWalletRepository: paymentWalletRepository,
		blockStateRepo:          blockStateRepo,
		cacheRepo:               cacheRepo,
		config:                  config,
	}
}

// CreatePaymentOrders creates new payment orders from the given payloads.
// It first attempts to retrieve the latest block from cache. If not found in cache, it fetches it from the database.
// Each payload is transformed into a PaymentOrder model, setting the block height to the latest block.
func (u *paymentOrderUCase) CreatePaymentOrders(
	ctx context.Context,
	payloads []dto.PaymentOrderPayloadDTO,
	expiredOrderTime time.Duration,
) ([]dto.CreatedPaymentOrderDTO, error) {
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

	var response []dto.CreatedPaymentOrderDTO

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
				Wallet:      *assignWallet,
				Transferred: "0", // Default transferred status
				RequestID:   payload.RequestID,
				Symbol:      payload.Symbol,
				BlockHeight: latestBlock,       // Use the latest block height
				Status:      constants.Pending, // Set order status to pending
				ExpiredTime: time.Now().UTC().Add(expiredOrderTime),
			}
			orders = append(orders, order)

			orderDTO := order.ToCreatedPaymentOrderDTO()
			orderDTO.PaymentAddress = assignWallet.Address
			response = append(response, orderDTO)
		}

		// Save the created payment orders to the database within the transaction
		orders, err = u.paymentOrderRepository.CreatePaymentOrders(ctx, orders)
		if err != nil {
			return fmt.Errorf("failed to create payment orders: %w", err)
		}

		// Mappping order id and sign payload
		response, err = u.mapOrderIDsAndSignPayloads(orders)
		if err != nil {
			return err
		}

		return nil
	})
	// Check if there was any error within the transaction
	if err != nil {
		return nil, err
	}

	return response, nil
}

// mapOrderIDsAndSignPayloads maps order IDs and signs payloads.
func (u *paymentOrderUCase) mapOrderIDsAndSignPayloads(
	orders []model.PaymentOrder,
) ([]dto.CreatedPaymentOrderDTO, error) {
	var response []dto.CreatedPaymentOrderDTO

	for _, order := range orders {
		orderDTO := order.ToCreatedPaymentOrderDTO()
		orderDTO.PaymentAddress = order.Wallet.Address

		// Decrypt the private key
		privKeyHex, err := utils.Decrypt(order.Wallet.PrivateKey, u.config.GetEncryptionKey())
		if err != nil {
			return nil, fmt.Errorf("error decrypting private key: %w", err)
		}

		// Convert to *ecdsa.PrivateKey
		privateKey, err := internalutils.PrivateKeyFromHex(privKeyHex)
		if err != nil {
			return nil, fmt.Errorf("error converting private key from hex: %w", err)
		}

		signature, err := utils.SignMessage(privateKey, []byte(orderDTO.RequestID+orderDTO.Amount+orderDTO.Symbol))
		if err != nil {
			return nil, fmt.Errorf("error signing message: %w", err)
		}

		orderDTO.Signature = signature
		response = append(response, orderDTO)
	}
	return response, nil
}

func (u *paymentOrderUCase) UpdateExpiredOrdersToFailed(ctx context.Context) error {
	return u.paymentOrderRepository.UpdateExpiredOrdersToFailed(ctx)
}

func (u *paymentOrderUCase) UpdateActiveOrdersToExpired(ctx context.Context) error {
	return u.paymentOrderRepository.UpdateActiveOrdersToExpired(ctx)
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

func (u *paymentOrderUCase) GetActivePaymentOrders(ctx context.Context, limit, offset int) ([]dto.PaymentOrderDTO, error) {
	orders, err := u.paymentOrderRepository.GetActivePaymentOrders(ctx, limit, offset)
	if err != nil {
		return nil, err
	}
	var orderDtos []dto.PaymentOrderDTO
	for _, order := range orders {
		orderDto := order.ToDto()
		orderDtos = append(orderDtos, orderDto)
	}
	return orderDtos, nil
}

func (u *paymentOrderUCase) GetPaymentOrderHistories(
	ctx context.Context,
	requestIDs []string,
	status *string,
	page, size int,
) (dto.PaginationDTOResponse, error) {
	// Setup pagination variables
	limit := size + 1 // Fetch one extra record to determine if there's a next page
	offset := (page - 1) * size

	// Fetch the orders with event histories from the repository
	orders, err := u.paymentOrderRepository.GetPaymentOrderHistories(ctx, limit, offset, requestIDs, status)
	if err != nil {
		return dto.PaginationDTOResponse{}, err
	}

	var orderHistoriesDTO []interface{}

	// Map orders to DTOs, limiting to requested page size
	for i, order := range orders {
		if i >= size { // Stop if we reach the requested page size
			break
		}
		orderDTO := dto.PaymentOrderDTOResponse{
			RequestID:      order.RequestID,
			Amount:         order.Amount,
			Transferred:    order.Transferred,
			Status:         order.Status,
			SucceededAt:    order.SucceededAt,
			EventHistories: mapEventHistoriesToDTO(order.PaymentEventHistories),
		}
		orderHistoriesDTO = append(orderHistoriesDTO, orderDTO)
	}

	// Determine if there's a next page
	nextPage := page
	if len(orders) > size {
		nextPage += 1
	}

	// Return the response DTO
	return dto.PaginationDTOResponse{
		NextPage: nextPage,
		Page:     page,
		Size:     size,
		Data:     orderHistoriesDTO,
	}, nil
}

// Helper function to map PaymentEventHistory to PaymentHistoryDTO
func mapEventHistoriesToDTO(eventHistories []model.PaymentEventHistory) []dto.PaymentHistoryDTO {
	eventHistoriesDTO := make([]dto.PaymentHistoryDTO, len(eventHistories))
	for i, eventHistory := range eventHistories {
		eventHistoriesDTO[i] = dto.PaymentHistoryDTO{
			TransactionHash: eventHistory.TransactionHash,
			FromAddress:     eventHistory.FromAddress,
			ToAddress:       eventHistory.ToAddress,
			Amount:          eventHistory.Amount,
			TokenSymbol:     eventHistory.TokenSymbol,
			CreatedAt:       eventHistory.CreatedAt,
		}
	}
	return eventHistoriesDTO
}
