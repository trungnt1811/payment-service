package interfaces

import (
	"context"
	"time"

	"github.com/genefriendway/onchain-handler/internal/dto"
	"github.com/genefriendway/onchain-handler/internal/model"
)

type PaymentOrderRepository interface {
	CreatePaymentOrders(ctx context.Context, orders []model.PaymentOrder) ([]model.PaymentOrder, error)
	GetActivePaymentOrders(ctx context.Context, limit, offset int) ([]model.PaymentOrder, error)
	UpdatePaymentOrder(
		ctx context.Context,
		orderID uint64,
		status, transferredAmount string,
		walletStatus bool,
		blockHeight uint64,
	) error
	BatchUpdateOrderStatuses(ctx context.Context, orderIDs []uint64, newStatuses []string) error
	GetExpiredPaymentOrders(ctx context.Context) ([]model.PaymentOrder, error)
	UpdateExpiredOrdersToFailed(ctx context.Context) error
}

type PaymentOrderUCase interface {
	CreatePaymentOrders(
		ctx context.Context,
		payloads []dto.PaymentOrderPayloadDTO,
		expiredOrderTime time.Duration,
	) ([]dto.PaymentOrderDTO, error)
	UpdateExpiredOrdersToFailed(ctx context.Context) error
	GetExpiredPaymentOrders(ctx context.Context) ([]model.PaymentOrder, error)
	UpdatePaymentOrder(
		ctx context.Context,
		orderID uint64,
		status, transferredAmount string,
		walletStatus bool,
		blockHeight uint64,
	) error
	BatchUpdateOrderStatuses(ctx context.Context, orders []dto.PaymentOrderDTO) error
	GetActivePaymentOrders(ctx context.Context, limit, offset int) ([]dto.PaymentOrderDTO, error)
}
