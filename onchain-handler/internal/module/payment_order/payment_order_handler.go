package payment_order

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/genefriendway/onchain-handler/conf"
	"github.com/genefriendway/onchain-handler/internal/dto"
	"github.com/genefriendway/onchain-handler/internal/interfaces"
	"github.com/genefriendway/onchain-handler/internal/utils/log"
)

type paymentOrderHandler struct {
	ucase  interfaces.PaymentOrderUCase
	config *conf.Configuration
}

func NewPaymentOrderHandler(
	ucase interfaces.PaymentOrderUCase, config *conf.Configuration,
) *paymentOrderHandler {
	return &paymentOrderHandler{
		ucase:  ucase,
		config: config,
	}
}

// CreateOrders creates payment orders based on the input payload.
// @Summary Create payment orders
// @Description This endpoint allows creating payment orders for users.
// @Tags payment-order
// @Accept json
// @Produce json
// @Param payload body []dto.PaymentOrderPayloadDTO true "List of payment orders. Each order must include user ID and amount."
// @Success 200 {object} map[string]interface{} "Success response: {\"success\": true, \"data\": []interface{}}"
// @Failure 400 {object} map[string]interface{} "Invalid payload"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/payment-orders [post]
func (h *paymentOrderHandler) CreateOrders(ctx *gin.Context) {
	var req []dto.PaymentOrderPayloadDTO

	// Parse and validate the request payload
	if err := ctx.ShouldBindJSON(&req); err != nil {
		log.LG.Errorf("Invalid payload: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid payload",
			"details": err.Error(),
		})
		return
	}

	// Validate the amount field for each payment order
	for _, order := range req {
		// Try to convert the amount string to a float
		if _, err := strconv.ParseFloat(order.Amount, 64); err != nil {
			log.LG.Errorf("Invalid amount for user %d: %v", order.UserID, err)
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error":   "Invalid amount",
				"details": "Amount must be a valid number",
			})
			return
		}
	}

	// Call the use case to create the payment orders
	response, err := h.ucase.CreatePaymentOrders(ctx, req, h.config.GetExpiredOrderTime())
	if err != nil {
		log.LG.Errorf("Failed to create payment orders: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to create payment orders",
			"details": err.Error(),
		})
		return
	}

	// Respond with success and response data
	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
	})
}
