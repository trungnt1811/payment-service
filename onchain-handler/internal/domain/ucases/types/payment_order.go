package types

import (
	"context"
	"time"

	"github.com/genefriendway/onchain-handler/constants"
	"github.com/genefriendway/onchain-handler/internal/delivery/dto"
)

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
	GetActivePaymentOrders(ctx context.Context) ([]dto.PaymentOrderDTO, error)
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
