package ucases

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/genefriendway/onchain-handler/conf"
	"github.com/genefriendway/onchain-handler/constants"
	"github.com/genefriendway/onchain-handler/infra/caching"
	infrainterfaces "github.com/genefriendway/onchain-handler/infra/interfaces"
	"github.com/genefriendway/onchain-handler/internal/domain"
	"github.com/genefriendway/onchain-handler/internal/dto"
	"github.com/genefriendway/onchain-handler/internal/interfaces"
	"github.com/genefriendway/onchain-handler/pkg/crypto"
)

type paymentOrderUCase struct {
	db                      *gorm.DB                           // Gorm DB instance for transaction handling
	paymentOrderRepository  interfaces.PaymentOrderRepository  // Repository to interact with payment orders
	paymentWalletRepository interfaces.PaymentWalletRepository // Repository to interact with payment wallets
	blockStateRepo          interfaces.BlockStateRepository    // Repository to fetch the latest block state
	cacheRepo               infrainterfaces.CacheRepository    // Cache repository to retrieve and store block information
	config                  *conf.Configuration
}

// NewPaymentOrderUCase constructs a new paymentOrderUCase with the provided dependencies.
func NewPaymentOrderUCase(
	db *gorm.DB, // Add DB to the constructor
	paymentOrderRepository interfaces.PaymentOrderRepository,
	paymentWalletRepository interfaces.PaymentWalletRepository,
	blockStateRepo interfaces.BlockStateRepository,
	cacheRepo infrainterfaces.CacheRepository,
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
	// Group payloads by network
	networkPayloads := make(map[string][]dto.PaymentOrderPayloadDTO)
	for _, payload := range payloads {
		networkPayloads[payload.Network] = append(networkPayloads[payload.Network], payload)
	}

	var response []dto.CreatedPaymentOrderDTO

	// Process each network's payloads
	for network, groupedPayloads := range networkPayloads {
		cacheKey := &caching.Keyer{Raw: constants.LatestBlockCacheKey + network} // Cache key for the latest block

		// Try to retrieve the latest block from cache
		var latestBlock uint64
		err := u.cacheRepo.RetrieveItem(cacheKey, &latestBlock)
		if err != nil {
			// If cache is empty, load from the database
			latestBlock, err = u.blockStateRepo.GetLatestBlock(ctx, network)
			if err != nil {
				return nil, fmt.Errorf("failed to retrieve latest block from database: %w", err)
			}
		}

		// Track claimed wallet IDs to roll back if needed
		var claimedWalletIDs []uint64

		// Begin a new transaction
		err = u.db.Transaction(func(tx *gorm.DB) error {
			var orders []domain.PaymentOrder

			// Convert each payload to a PaymentOrder model, associating it with the latest block
			for _, payload := range groupedPayloads {
				// Try to claim an available wallet
				assignWallet, err := u.paymentWalletRepository.ClaimFirstAvailableWallet(ctx)
				if err != nil {
					return fmt.Errorf("failed to claim available wallet: %w", err)
				}

				// Track claimed wallet ID for potential rollback
				claimedWalletIDs = append(claimedWalletIDs, assignWallet.ID)

				// Create the PaymentOrder model for the current payload
				order := domain.PaymentOrder{
					Amount:      payload.Amount,
					WalletID:    assignWallet.ID, // Associate the order with the claimed wallet
					Wallet:      *assignWallet,
					Transferred: "0", // Default transferred status
					RequestID:   payload.RequestID,
					Symbol:      payload.Symbol,
					Network:     payload.Network,
					WebhookURL:  payload.WebhookURL,
					BlockHeight: latestBlock,       // Use the latest block height
					Status:      constants.Pending, // Set order status to pending
					ExpiredTime: time.Now().UTC().Add(expiredOrderTime),
				}
				orders = append(orders, order)
			}

			// Save the created payment orders to the database within the transaction
			orders, err = u.paymentOrderRepository.CreatePaymentOrders(ctx, orders)
			if err != nil {
				return fmt.Errorf("failed to create payment orders: %w", err)
			}

			// Mappping order id and sign payload
			responseWithSignature, err := u.mapOrderIDsAndSignPayloads(orders)
			if err != nil {
				return err
			}
			response = append(response, responseWithSignature...)
			return nil
		})
		// Check if there was an error within the transaction
		if err != nil {
			// Rollback claimed wallets
			rollbackErr := u.paymentWalletRepository.BatchReleaseWallets(ctx, claimedWalletIDs)
			if rollbackErr != nil {
				return nil, fmt.Errorf("transaction failed, and rollback also failed: %w", rollbackErr)
			}
			return nil, err
		}
	}

	return response, nil
}

// mapOrderIDsAndSignPayloads maps order IDs and signs payloads.
func (u *paymentOrderUCase) mapOrderIDsAndSignPayloads(
	orders []domain.PaymentOrder,
) ([]dto.CreatedPaymentOrderDTO, error) {
	var response []dto.CreatedPaymentOrderDTO

	for _, order := range orders {
		orderDTO := order.ToCreatedPaymentOrderDTO()
		orderDTO.PaymentAddress = order.Wallet.Address
		orderDTO.Expired = uint64(order.ExpiredTime.Unix())

		// Get private key from wallet id
		_, privateKey, err := crypto.GenerateAccount(
			u.config.Wallet.Mnemonic,
			u.config.Wallet.Passphrase,
			u.config.Wallet.Salt,
			constants.PaymentWallet,
			order.Wallet.ID,
		)
		if err != nil {
			return nil, fmt.Errorf("error get payment wallet private key: %w", err)
		}

		signature, err := crypto.SignMessage(privateKey, []byte(orderDTO.RequestID+orderDTO.Amount+orderDTO.Symbol))
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
	_, err := u.paymentOrderRepository.UpdateActiveOrdersToExpired(ctx)
	return err
}

func (u *paymentOrderUCase) GetExpiredPaymentOrders(ctx context.Context, network constants.NetworkType) ([]dto.PaymentOrderDTO, error) {
	var orderDtos []dto.PaymentOrderDTO
	expiredOrders, err := u.paymentOrderRepository.GetExpiredPaymentOrders(ctx, string(network))
	if err != nil {
		return nil, err
	}
	for _, order := range expiredOrders {
		orderDto := order.ToDto()
		orderDtos = append(orderDtos, orderDto)
	}
	return orderDtos, nil
}

func (u *paymentOrderUCase) UpdatePaymentOrder(
	ctx context.Context,
	orderID uint64,
	blockHeight *uint64,
	status, transferredAmount, network *string,
) error {
	// Retrieve the payment order from the repository
	order, err := u.paymentOrderRepository.GetPaymentOrderByID(ctx, orderID)
	if err != nil {
		return fmt.Errorf("failed to retrieve payment order with id %d: %w", orderID, err)
	}

	// Update the fields of the payment order
	if status != nil {
		order.Status = *status
		// Update the succeeded_at timestamp if the status is "Success"
		if *status == constants.Success {
			order.SucceededAt = time.Now().UTC()
		}
	}

	if transferredAmount != nil {
		order.Transferred = *transferredAmount
	}

	if network != nil {
		if order.Status == constants.Pending {
			order.Network = *network
		} else {
			return fmt.Errorf("failed to update payment order with id %d: order status is not PENDING", orderID)
		}
	}

	if blockHeight != nil {
		order.BlockHeight = *blockHeight
	}

	// Persist the updated payment order to the database
	if err := u.paymentOrderRepository.UpdatePaymentOrder(ctx, order); err != nil {
		return fmt.Errorf("failed to update payment order with id %d: %w", orderID, err)
	}

	return nil
}

func (u *paymentOrderUCase) UpdateOrderStatus(ctx context.Context, orderID uint64, newStatus string) error {
	return u.paymentOrderRepository.UpdateOrderStatus(ctx, orderID, newStatus)
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

func (u *paymentOrderUCase) BatchUpdateOrderBlockHeights(ctx context.Context, orders []dto.PaymentOrderDTO) error {
	var orderIDs []uint64
	var blockHeights []uint64
	// Parse orders to extract orderIDs and newStatuses
	for _, order := range orders {
		orderIDs = append(orderIDs, order.ID)
		blockHeights = append(blockHeights, order.BlockHeight)
	}

	return u.paymentOrderRepository.BatchUpdateOrderBlockHeights(ctx, orderIDs, blockHeights)
}

func (u *paymentOrderUCase) GetActivePaymentOrdersOnAvax(ctx context.Context, limit, offset int) ([]dto.PaymentOrderDTO, error) {
	orders, err := u.paymentOrderRepository.GetActivePaymentOrders(ctx, limit, offset, string(constants.AvaxCChain))
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

func (u *paymentOrderUCase) GetActivePaymentOrdersOnBsc(ctx context.Context, limit, offset int) ([]dto.PaymentOrderDTO, error) {
	orders, err := u.paymentOrderRepository.GetActivePaymentOrders(ctx, limit, offset, string(constants.Bsc))
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

func (u *paymentOrderUCase) GetPaymentOrders(
	ctx context.Context,
	status, orderBy *string,
	orderDirection constants.OrderDirection,
	page, size int,
) (dto.PaginationDTOResponse, error) {
	// Setup pagination variables
	limit := size + 1 // Fetch one extra record to determine if there's a next page
	offset := (page - 1) * size

	// Fetch the orders with event histories from the repository
	orders, err := u.paymentOrderRepository.GetPaymentOrders(ctx, limit, offset, status, orderBy, orderDirection)
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
			ID:          order.ID,
			RequestID:   order.RequestID,
			Network:     order.Network,
			Amount:      order.Amount,
			Symbol:      order.Symbol,
			Transferred: order.Transferred,
			Status:      order.Status,
		}
		if order.Status == constants.Success {
			orderDTO.SucceededAt = &order.SucceededAt
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

func (u *paymentOrderUCase) GetPaymentOrderByID(ctx context.Context, id uint64) (dto.PaymentOrderDTOResponse, error) {
	order, err := u.paymentOrderRepository.GetPaymentOrderByID(ctx, id)
	if err != nil {
		return dto.PaymentOrderDTOResponse{}, err
	}
	orderDTO := dto.PaymentOrderDTOResponse{
		ID:             order.ID,
		RequestID:      order.RequestID,
		Network:        order.Network,
		Amount:         order.Amount,
		Transferred:    order.Transferred,
		Status:         order.Status,
		WebhookURL:     order.WebhookURL,
		Symbol:         order.Symbol,
		WalletAddress:  &order.Wallet.Address,
		Expired:        uint64(order.ExpiredTime.Unix()),
		EventHistories: mapEventHistoriesToDTO(order.PaymentEventHistories),
	}
	if order.Status == constants.Success {
		orderDTO.SucceededAt = &order.SucceededAt
	}
	return orderDTO, nil
}

// Helper function to map PaymentEventHistory to PaymentHistoryDTO
func mapEventHistoriesToDTO(eventHistories []domain.PaymentEventHistory) []dto.PaymentHistoryDTO {
	eventHistoriesDTO := make([]dto.PaymentHistoryDTO, len(eventHistories))
	for i, eventHistory := range eventHistories {
		eventHistoriesDTO[i] = dto.PaymentHistoryDTO{
			TransactionHash: eventHistory.TransactionHash,
			FromAddress:     eventHistory.FromAddress,
			ToAddress:       eventHistory.ToAddress,
			Amount:          eventHistory.Amount,
			Network:         eventHistory.Network,
			TokenSymbol:     eventHistory.TokenSymbol,
			CreatedAt:       eventHistory.CreatedAt,
		}
	}
	return eventHistoriesDTO
}
