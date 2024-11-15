package interfaces

import (
	"context"
	"time"

	"github.com/genefriendway/onchain-handler/constants"
	"github.com/genefriendway/onchain-handler/internal/dto"
	"github.com/genefriendway/onchain-handler/internal/model"
)

type PaymentOrderRepository interface {
	CreatePaymentOrders(ctx context.Context, orders []model.PaymentOrder) ([]model.PaymentOrder, error)
	GetActivePaymentOrders(ctx context.Context, limit, offset int, network string) ([]model.PaymentOrder, error)
	UpdatePaymentOrder(
		ctx context.Context,
		orderID uint64,
		status, transferredAmount string,
		walletStatus bool,
		blockHeight uint64,
	) error
	BatchUpdateOrderStatuses(ctx context.Context, orderIDs []uint64, newStatuses []string) error
	BatchUpdateOrderBlockHeights(ctx context.Context, orderIDs, blockHeights []uint64) error
	GetExpiredPaymentOrders(ctx context.Context, network string) ([]model.PaymentOrder, error)
	UpdateExpiredOrdersToFailed(ctx context.Context) error
	UpdateActiveOrdersToExpired(ctx context.Context) error
	GetPaymentOrderHistories(ctx context.Context, limit, offset int, requestIDs []string, status *string) (
		[]model.PaymentOrder, error,
	)
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
		status, transferredAmount string,
		walletStatus bool,
		blockHeight uint64,
	) error
	BatchUpdateOrderStatuses(ctx context.Context, orders []dto.PaymentOrderDTO) error
	BatchUpdateOrderBlockHeights(ctx context.Context, orders []dto.PaymentOrderDTO) error
	GetActivePaymentOrdersOnAvax(ctx context.Context, limit, offset int) ([]dto.PaymentOrderDTO, error)
	GetActivePaymentOrdersOnBsc(ctx context.Context, limit, offset int) ([]dto.PaymentOrderDTO, error)
	GetPaymentOrderHistories(ctx context.Context, requestIDs []string, status *string, page, size int) (
		dto.PaginationDTOResponse, error,
	)
}
