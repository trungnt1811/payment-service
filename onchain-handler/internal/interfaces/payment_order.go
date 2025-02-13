package interfaces

import (
	"context"
	"time"

	"gorm.io/gorm"

	"github.com/genefriendway/onchain-handler/constants"
	"github.com/genefriendway/onchain-handler/internal/domain/dto"
	"github.com/genefriendway/onchain-handler/internal/domain/entities"
)

type PaymentOrderRepository interface {
	CreatePaymentOrders(tx *gorm.DB,
		ctx context.Context,
		orders []entities.PaymentOrder,
		vendorID string,
	) ([]entities.PaymentOrder, error)
	GetActivePaymentOrders(ctx context.Context, limit, offset int, network *string) ([]entities.PaymentOrder, error)
	UpdatePaymentOrder(ctx context.Context, order *entities.PaymentOrder) error
	UpdateOrderNetwork(ctx context.Context, requestID, network string, blockHeight uint64) error
	BatchUpdateOrdersToExpired(ctx context.Context, orderIDs []uint64) error
	BatchUpdateOrderBlockHeights(ctx context.Context, orderIDs, blockHeights []uint64) error
	GetExpiredPaymentOrders(ctx context.Context, network string) ([]entities.PaymentOrder, error)
	UpdateExpiredOrdersToFailed(ctx context.Context) ([]uint64, error)
	UpdateActiveOrdersToExpired(ctx context.Context) ([]uint64, error)
	GetPaymentOrders(
		ctx context.Context,
		limit, offset int,
		vendorID string,
		requestIDs []string,
		status, orderBy, fromAddress, network *string,
		orderDirection constants.OrderDirection,
		startTime, endTime *time.Time,
		timeFilterField *string,
	) ([]entities.PaymentOrder, error)
	GetPaymentOrderByID(ctx context.Context, id uint64) (*entities.PaymentOrder, error)
	GetPaymentOrdersByIDs(ctx context.Context, ids []uint64) ([]entities.PaymentOrder, error)
	GetPaymentOrderByRequestID(ctx context.Context, requestID string) (*entities.PaymentOrder, error)
	GetPaymentOrderIDByRequestID(ctx context.Context, requestID string) (uint64, error)
}

type PaymentOrderUCase interface {
	CreatePaymentOrders(
		ctx context.Context,
		payloads []dto.PaymentOrderPayloadDTO,
		vendorID string,
		expiredOrderTime time.Duration,
	) ([]dto.CreatedPaymentOrderDTO, error)
	UpdateExpiredOrdersToFailed(ctx context.Context) ([]uint64, error)
	UpdateActiveOrdersToExpired(ctx context.Context) ([]uint64, error)
	GetExpiredPaymentOrders(ctx context.Context, network constants.NetworkType) ([]dto.PaymentOrderDTO, error)
	UpdatePaymentOrder(
		ctx context.Context,
		orderID uint64,
		blockHeight, upcomingBlockHeight *uint64,
		status, transferredAmount, network *string,
	) error
	UpdateOrderNetwork(ctx context.Context, requestID string, network constants.NetworkType) error
	BatchUpdateOrdersToExpired(ctx context.Context, orderIDs []uint64) error
	BatchUpdateOrderBlockHeights(ctx context.Context, orders []dto.PaymentOrderDTO) error
	GetActivePaymentOrders(ctx context.Context, limit, offset int) ([]dto.PaymentOrderDTO, error)
	GetPaymentOrders(
		ctx context.Context,
		vendorID string,
		requestIDs []string,
		status, orderBy, fromAddress, network *string,
		orderDirection constants.OrderDirection,
		startTime, endTime *time.Time,
		timeFilterField *string,
		page, size int,
	) (dto.PaginationDTOResponse, error)
	GetPaymentOrderByID(ctx context.Context, id uint64) (dto.PaymentOrderDTOResponse, error)
	GetPaymentOrdersByIDs(ctx context.Context, ids []uint64) ([]dto.PaymentOrderDTOResponse, error)
	GetPaymentOrderByRequestID(ctx context.Context, requestID string) (dto.PaymentOrderDTOResponse, error)
}
