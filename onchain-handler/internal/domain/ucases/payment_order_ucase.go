package ucases

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/genefriendway/onchain-handler/constants"
	cachetypes "github.com/genefriendway/onchain-handler/internal/adapters/cache/types"
	settypes "github.com/genefriendway/onchain-handler/internal/adapters/orderset/types"
	repotypes "github.com/genefriendway/onchain-handler/internal/adapters/repositories/types"
	"github.com/genefriendway/onchain-handler/internal/delivery/dto"
	"github.com/genefriendway/onchain-handler/internal/domain/entities"
	ucasetypes "github.com/genefriendway/onchain-handler/internal/domain/ucases/types"
	"github.com/genefriendway/onchain-handler/pkg/blockchain"
	"github.com/genefriendway/onchain-handler/pkg/logger"
	"github.com/genefriendway/onchain-handler/pkg/utils"
)

type paymentOrderUCase struct {
	db                          *gorm.DB                              // Gorm DB instance for transaction handling
	paymentOrderRepository      repotypes.PaymentOrderRepository      // Repository to interact with payment orders
	paymentWalletRepository     repotypes.PaymentWalletRepository     // Repository to interact with payment wallets
	blockStateRepo              repotypes.BlockStateRepository        // Repository to fetch the latest block state
	paymentStatisticsRepository repotypes.PaymentStatisticsRepository // Repository to interact with payment statistics
	cacheRepo                   cachetypes.CacheRepository            // Cache repository to retrieve and store block information
	paymentOrderSet             settypes.Set[dto.PaymentOrderDTO]     // Payment order set
}

// NewPaymentOrderUCase constructs a new paymentOrderUCase with the provided dependencies.
func NewPaymentOrderUCase(
	db *gorm.DB, // Add DB to the constructor
	paymentOrderRepository repotypes.PaymentOrderRepository,
	paymentWalletRepository repotypes.PaymentWalletRepository,
	blockStateRepo repotypes.BlockStateRepository,
	paymentStatisticsRepository repotypes.PaymentStatisticsRepository,
	cacheRepo cachetypes.CacheRepository,
	paymentOrderSet settypes.Set[dto.PaymentOrderDTO],
) ucasetypes.PaymentOrderUCase {
	return &paymentOrderUCase{
		db:                          db,
		paymentOrderRepository:      paymentOrderRepository,
		paymentWalletRepository:     paymentWalletRepository,
		blockStateRepo:              blockStateRepo,
		paymentStatisticsRepository: paymentStatisticsRepository,
		cacheRepo:                   cacheRepo,
		paymentOrderSet:             paymentOrderSet,
	}
}

// CreatePaymentOrders creates new payment orders from the given payloads.
// It first attempts to retrieve the latest block from cache. If not found in cache, it fetches it from the database.
// Each payload is transformed into a PaymentOrder model, setting the block height to the latest block.
func (u *paymentOrderUCase) CreatePaymentOrders(
	ctx context.Context,
	payloads []dto.PaymentOrderPayloadDTO,
	vendorID string,
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
		// Retrieve the latest block
		cacheKey := &cachetypes.Keyer{Raw: constants.LatestBlockCacheKey + network}
		var latestBlock uint64

		// Attempt to retrieve the latest block from cache
		err := u.cacheRepo.RetrieveItem(cacheKey, &latestBlock)
		if err != nil {
			// If not in cache, retrieve from the database
			latestBlock, err = u.blockStateRepo.GetLatestBlock(ctx, network)
			if err != nil {
				return nil, fmt.Errorf("failed to retrieve latest block from database: %w", err)
			}
		}

		// Begin transaction and process orders
		err = u.processOrderPayloads(ctx, groupedPayloads, latestBlock, vendorID, expiredOrderTime, &response)
		if err != nil {
			return nil, err
		}

		// Update payment statistics
		err = u.updatePaymentStatistics(ctx, groupedPayloads, vendorID)
		if err != nil {
			return nil, err
		}
	}

	return response, nil
}

func (u *paymentOrderUCase) updatePaymentStatistics(
	ctx context.Context,
	payloads []dto.PaymentOrderPayloadDTO,
	vendorID string,
) error {
	granularity := constants.Daily
	periodStart := utils.GetPeriodStart(granularity, time.Now())

	for index, payload := range payloads {
		err := u.paymentStatisticsRepository.IncrementStatistics(
			ctx,
			granularity,
			periodStart.UTC(),
			&payloads[index].Amount,
			nil, // Transferred value is not updated here
			payload.Symbol,
			vendorID,
		)
		if err != nil {
			return fmt.Errorf("failed to increment payment statistics: %w", err)
		}
	}

	return nil
}

func (u *paymentOrderUCase) processOrderPayloads(
	ctx context.Context,
	payloads []dto.PaymentOrderPayloadDTO,
	latestBlock uint64,
	vendorID string,
	expiredOrderTime time.Duration,
	response *[]dto.CreatedPaymentOrderDTO,
) error {
	var claimedWalletIDs []uint64
	var createdOrders []entities.PaymentOrder

	// Begin transaction
	err := u.db.Transaction(func(tx *gorm.DB) error {
		var orders []entities.PaymentOrder

		for _, payload := range payloads {
			// Step 1: Claim an available wallet inside the transaction
			wallet, err := u.paymentWalletRepository.ClaimFirstAvailableWallet(tx, ctx)
			if err != nil {
				return fmt.Errorf("failed to claim available wallet: %w", err)
			}

			// Track claimed wallet IDs
			claimedWalletIDs = append(claimedWalletIDs, wallet.ID)

			// Step 2: Create a new payment order
			order := entities.PaymentOrder{
				Amount:      payload.Amount,
				WalletID:    wallet.ID,
				Wallet:      *wallet,
				Transferred: "0",
				RequestID:   payload.RequestID,
				Symbol:      payload.Symbol,
				Network:     payload.Network,
				WebhookURL:  payload.WebhookURL,
				BlockHeight: latestBlock,
				Status:      constants.Pending,
				ExpiredTime: time.Now().UTC().Add(expiredOrderTime),
			}
			orders = append(orders, order)
		}

		// Step 3: Save orders within the transaction
		var err error
		createdOrders, err = u.paymentOrderRepository.CreatePaymentOrders(tx, ctx, orders, vendorID)
		if err != nil {
			// Rollback claimed wallets inside the transaction
			if rollbackErr := u.paymentWalletRepository.ReleaseWalletsByIDs(tx, claimedWalletIDs); rollbackErr != nil {
				return fmt.Errorf("failed to rollback wallets: %w", rollbackErr)
			}
			return fmt.Errorf("failed to create payment orders: %w", err)
		}

		return nil // Commit transaction
	})
	if err != nil {
		return err // Return error if transaction fails
	}

	// Step 4: Add orders to the payment order set (AFTER transaction commit)
	for _, order := range createdOrders {
		if addErr := u.paymentOrderSet.Add(order.ToDto()); addErr != nil {
			// Log error instead of failing the whole process
			logger.GetLogger().Errorf("Failed to add order to payment order set: %v", addErr)
		}
	}

	// Step 5: Map order IDs
	responseWithIDs, err := u.mapOrderIDs(createdOrders)
	if err != nil {
		return fmt.Errorf("failed to map order IDs and sign payloads: %w", err)
	}

	*response = append(*response, responseWithIDs...)
	return nil
}

// mapOrderIDs maps order IDs.
func (u *paymentOrderUCase) mapOrderIDs(
	orders []entities.PaymentOrder,
) ([]dto.CreatedPaymentOrderDTO, error) {
	var response []dto.CreatedPaymentOrderDTO

	for _, order := range orders {
		orderDTO := order.ToCreatedPaymentOrderDTO()
		orderDTO.PaymentAddress = order.Wallet.Address
		orderDTO.Expired = uint64(order.ExpiredTime.Unix())
		response = append(response, orderDTO)
	}
	return response, nil
}

func (u *paymentOrderUCase) UpdateExpiredOrdersToFailed(ctx context.Context) ([]uint64, error) {
	return u.paymentOrderRepository.UpdateExpiredOrdersToFailed(ctx)
}

func (u *paymentOrderUCase) UpdateActiveOrdersToExpired(ctx context.Context) ([]uint64, error) {
	return u.paymentOrderRepository.UpdateActiveOrdersToExpired(ctx)
}

func (u *paymentOrderUCase) GetExpiredPaymentOrders(ctx context.Context, network constants.NetworkType) ([]dto.PaymentOrderDTO, error) {
	var orderDtos []dto.PaymentOrderDTO
	expiredOrders, err := u.paymentOrderRepository.GetExpiredPaymentOrders(ctx, network.String())
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
	blockHeight, upcomingBlockHeight *uint64,
	status, transferredAmount, network *string,
) error {
	return u.paymentOrderRepository.UpdatePaymentOrder(ctx, orderID, func(order *entities.PaymentOrder) error {
		if status != nil {
			order.Status = *status
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
				return fmt.Errorf("cannot update network: order status is not PENDING")
			}
		}

		if blockHeight != nil {
			order.BlockHeight = *blockHeight
		}

		if upcomingBlockHeight != nil {
			order.UpcomingBlockHeight = *upcomingBlockHeight
		}

		return nil
	})
}

func (u *paymentOrderUCase) UpdateOrderToSuccessAndReleaseWallet(
	ctx context.Context,
	orderID uint64,
) error {
	return u.paymentOrderRepository.UpdateOrderToSuccessAndReleaseWallet(
		ctx,
		orderID,
		time.Now().UTC(),
	)
}

func (u *paymentOrderUCase) UpdateOrderNetwork(ctx context.Context, requestID string, network constants.NetworkType) error {
	// Step 1: Retrieve the payment order by request ID
	order, err := u.paymentOrderRepository.GetPaymentOrderByRequestID(ctx, requestID)
	if err != nil {
		return fmt.Errorf("failed to retrieve payment order with request id %s: %w", requestID, err)
	}

	// Step 2: Check if the order is pending
	if order.Status != constants.Pending {
		return fmt.Errorf("failed to update payment order with id %d: order status is not PENDING", order.ID)
	}

	// Step 3: Fetch the latest block height for the given network
	latestBlock, err := blockchain.GetLatestBlockFromCache(ctx, network.String(), u.cacheRepo)
	if err != nil {
		return fmt.Errorf("failed to get latest block from cache: %w", err)
	}

	// Step 4: Update the order in the database
	err = u.paymentOrderRepository.UpdateOrderNetwork(ctx, requestID, network.String(), latestBlock)
	if err != nil {
		return fmt.Errorf("failed to update order network: %w", err)
	}

	// Step 5: Update order fields
	order.Network = network.String()
	order.BlockHeight = latestBlock

	// Step 6: Update the payment order set
	orderDTO := order.ToDto()
	key := orderDTO.PaymentAddress + "_" + orderDTO.Symbol
	err = u.paymentOrderSet.UpdateItem(key, orderDTO)
	if err != nil {
		logger.GetLogger().Errorf("Failed to update payment order in set: %v", err)
		return fmt.Errorf("failed to update payment order in memory: %w", err)
	}

	return nil
}

func (u *paymentOrderUCase) BatchUpdateOrdersToExpired(ctx context.Context, orderIDs []uint64) error {
	return u.paymentOrderRepository.BatchUpdateOrdersToExpired(ctx, orderIDs)
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

func (u *paymentOrderUCase) GetActivePaymentOrders(ctx context.Context) ([]dto.PaymentOrderDTO, error) {
	orders, err := u.paymentOrderRepository.GetActivePaymentOrders(ctx, nil)
	if err != nil {
		return nil, err
	}
	var orderDTOs []dto.PaymentOrderDTO
	for _, order := range orders {
		orderDTO := order.ToDto()
		orderDTOs = append(orderDTOs, orderDTO)
	}
	return orderDTOs, nil
}

func (u *paymentOrderUCase) GetPaymentOrders(
	ctx context.Context,
	vendorID string,
	requestIDs []string,
	status, orderBy, fromAddress, network *string,
	orderDirection constants.OrderDirection,
	startTime, endTime *time.Time,
	timeFilterField *string,
	page, size int,
) (dto.PaginationDTOResponse, error) {
	// Setup pagination variables
	limit := size + 1 // Fetch one extra record to determine if there's a next page
	offset := (page - 1) * size

	// Fetch the orders with event histories from the repository
	orders, err := u.paymentOrderRepository.GetPaymentOrders(
		ctx, limit, offset, vendorID, requestIDs, status, orderBy, fromAddress, network, orderDirection, startTime, endTime, timeFilterField,
	)
	if err != nil {
		return dto.PaginationDTOResponse{}, err
	}

	var orderHistoriesDTO []any

	// Map orders to DTOs, limiting to requested page size
	for i, order := range orders {
		if i >= size { // Stop if we reach the requested page size
			break
		}
		orderHistoriesDTO = append(orderHistoriesDTO, mapOrderToDTO(order))
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
	// Fetch the order from the repository
	order, err := u.paymentOrderRepository.GetPaymentOrderByID(ctx, id)
	if err != nil {
		return dto.PaymentOrderDTOResponse{}, err
	}

	// Map the order to a DTO
	orderDTO := mapOrderToDTO(*order)

	return orderDTO, nil
}

func (u *paymentOrderUCase) GetPaymentOrdersByIDs(ctx context.Context, ids []uint64) ([]dto.PaymentOrderDTOResponse, error) {
	// Fetch orders from the repository
	orders, err := u.paymentOrderRepository.GetPaymentOrdersByIDs(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve payment orders: %w", err)
	}

	// Map orders to DTOs
	var orderDTOs []dto.PaymentOrderDTOResponse
	for _, order := range orders {
		orderDTOs = append(orderDTOs, mapOrderToDTO(order))
	}

	return orderDTOs, nil
}

func (u *paymentOrderUCase) GetPaymentOrderByRequestID(ctx context.Context, requestID string) (dto.PaymentOrderDTOResponse, error) {
	// Fetch the payment order by request ID
	order, err := u.paymentOrderRepository.GetPaymentOrderByRequestID(ctx, requestID)
	if err != nil {
		// Return a clean not found error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return dto.PaymentOrderDTOResponse{}, gorm.ErrRecordNotFound
		}
		return dto.PaymentOrderDTOResponse{}, fmt.Errorf("failed to retrieve payment order: %w", err)
	}

	// Map the order to a DTO
	orderDTO := mapOrderToDTO(*order)

	return orderDTO, nil
}

// Helper function to map PaymentOrder to PaymentOrderDTOResponse
func mapOrderToDTO(order entities.PaymentOrder) dto.PaymentOrderDTOResponse {
	dto := dto.PaymentOrderDTOResponse{
		ID:                  order.ID,
		RequestID:           order.RequestID,
		Network:             order.Network,
		Amount:              order.Amount,
		Transferred:         order.Transferred,
		Status:              order.Status,
		WebhookURL:          order.WebhookURL,
		Symbol:              order.Symbol,
		BlockHeight:         order.BlockHeight,
		UpcomingBlockHeight: order.UpcomingBlockHeight,
		PaymentAddress:      order.Wallet.Address,
		CreatedAt:           order.CreatedAt,
		Expired:             uint64(order.ExpiredTime.Unix()),
		EventHistories:      mapEventHistoriesToDTO(order.PaymentEventHistories),
	}
	if order.Status == constants.Success {
		dto.SucceededAt = &order.SucceededAt
	}
	return dto
}

// Helper function to map PaymentEventHistory to PaymentHistoryDTO
func mapEventHistoriesToDTO(eventHistories []entities.PaymentEventHistory) []dto.PaymentHistoryDTO {
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

func (u *paymentOrderUCase) ReleaseWalletsForSuccessfulOrders(ctx context.Context) error {
	err := u.paymentOrderRepository.ReleaseWalletsForSuccessfulOrders(ctx)
	if err != nil {
		logger.GetLogger().Errorf("Failed to release wallets for successful orders: %v", err)
		return err
	}

	logger.GetLogger().Info("Successfully released wallets for successful orders.")
	return nil
}

func (u *paymentOrderUCase) GetProcessingOrdersExpired(ctx context.Context, network constants.NetworkType) ([]dto.PaymentOrderDTOResponse, error) {
	orders, err := u.paymentOrderRepository.GetProcessingOrdersExpired(ctx, network.String())
	if err != nil {
		return nil, fmt.Errorf("failed to get expired processing orders: %w", err)
	}

	var result []dto.PaymentOrderDTOResponse
	for _, order := range orders {
		result = append(result, mapOrderToDTO(order))
	}
	return result, nil
}
