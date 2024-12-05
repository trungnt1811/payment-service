package interfaces

import (
	"context"
	"time"

	"github.com/genefriendway/onchain-handler/constants"
	"github.com/genefriendway/onchain-handler/internal/domain"
	"github.com/genefriendway/onchain-handler/internal/dto"
)

type PaymentOrderRepository interface {
	CreatePaymentOrders(ctx context.Context, orders []domain.PaymentOrder) ([]domain.PaymentOrder, error)
	GetActivePaymentOrders(ctx context.Context, limit, offset int, network *string) ([]domain.PaymentOrder, error)
	UpdatePaymentOrder(ctx context.Context, order *domain.PaymentOrder) error
	UpdateOrderStatus(ctx context.Context, orderID uint64, newStatus string) error
	UpdateOrderNetwork(ctx context.Context, orderID uint64, network string) error
	BatchUpdateOrderStatuses(ctx context.Context, orderIDs []uint64, newStatuses []string) error
	BatchUpdateOrderBlockHeights(ctx context.Context, orderIDs, blockHeights []uint64) error
	GetExpiredPaymentOrders(ctx context.Context, network string) ([]domain.PaymentOrder, error)
	UpdateExpiredOrdersToFailed(ctx context.Context) error
	UpdateActiveOrdersToExpired(ctx context.Context) ([]uint64, error)
	GetPaymentOrders(
		ctx context.Context,
		limit, offset int,
		status, orderBy *string,
		orderDirection constants.OrderDirection,
	) ([]domain.PaymentOrder, error)
	GetPaymentOrderByID(ctx context.Context, id uint64) (*domain.PaymentOrder, error)
	GetPaymentOrderByRequestID(ctx context.Context, requestID string) (*domain.PaymentOrder, error)
}

type PaymentOrderUCase interface {
	CreatePaymentOrders(
		ctx context.Context,
		payloads []dto.PaymentOrderPayloadDTO,
		expiredOrderTime time.Duration,
	) ([]dto.CreatedPaymentOrderDTO, error)
	UpdateExpiredOrdersToFailed(ctx context.Context) error
	UpdateActiveOrdersToExpired(ctx context.Context) error
	GetExpiredPaymentOrders(ctx context.Context, network constants.NetworkType) ([]dto.PaymentOrderDTO, error)
	UpdatePaymentOrder(
		ctx context.Context,
		orderID uint64,
		blockHeight *uint64,
		status, transferredAmount, network *string,
	) error
	UpdateOrderStatus(ctx context.Context, orderID uint64, newStatus string) error
	UpdateOrderNetwork(ctx context.Context, orderID uint64, network constants.NetworkType) error
	BatchUpdateOrderStatuses(ctx context.Context, orders []dto.PaymentOrderDTO) error
	BatchUpdateOrderBlockHeights(ctx context.Context, orders []dto.PaymentOrderDTO) error
	GetActivePaymentOrders(ctx context.Context, limit, offset int) ([]dto.PaymentOrderDTO, error)
	GetPaymentOrders(
		ctx context.Context,
		status, orderBy *string,
		orderDirection constants.OrderDirection,
		page, size int,
	) (dto.PaginationDTOResponse, error)
	GetPaymentOrderByID(ctx context.Context, id uint64) (dto.PaymentOrderDTOResponse, error)
	GetPaymentOrderByRequestID(ctx context.Context, requestID string) (dto.PaymentOrderDTOResponse, error)
}
