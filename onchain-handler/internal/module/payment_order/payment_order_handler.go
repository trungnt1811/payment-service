package payment_order

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/genefriendway/onchain-handler/conf"
	"github.com/genefriendway/onchain-handler/constants"
	"github.com/genefriendway/onchain-handler/internal/dto"
	"github.com/genefriendway/onchain-handler/internal/interfaces"
	"github.com/genefriendway/onchain-handler/log"
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
// @Param payload body []dto.PaymentOrderPayloadDTO true "List of payment orders. Each order must include request id, amount, symbol (USDT) and network (AVAX C-Chain)."
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

	// Validate each payment order
	for _, order := range req {
		if err := validatePaymentOrder(order); err != nil {
			log.LG.Errorf("Validation failed for request id %s: %v", order.RequestID, err)
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error":   "Validation error",
				"details": err.Error(),
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

// validatePaymentOrder performs validation checks on the payment order.
func validatePaymentOrder(order dto.PaymentOrderPayloadDTO) error {
	// Validate amount is a valid float
	if _, err := strconv.ParseFloat(order.Amount, 64); err != nil {
		return fmt.Errorf("invalid amount: %v", err)
	}

	// Validate symbol is USDT
	validSymbols := map[string]bool{
		constants.USDT: true,
		// constants.LP:   true,
	}
	if !validSymbols[order.Symbol] {
		return fmt.Errorf("invalid symbol: %s, must be USDT", order.Symbol)
	}

	// Validate network type is either BSC or AVAX C-Chain
	validNetworks := map[constants.NetworkType]bool{
		constants.Bsc:        true,
		constants.AvaxCChain: true,
	}
	networkType := constants.NetworkType(order.Network)
	if !validNetworks[networkType] {
		return fmt.Errorf("invalid network type: %s, must be BSC or AVAX C-Chain", order.Network)
	}

	return nil
}

// GetPaymentOrderHistories retrieves payment orders by request IDs and optionally filters by status.
// @Summary Retrieve payment order histories
// @Description This endpoint retrieves payment order histories based on request IDs and an optional status filter.
// @Tags payment-order
// @Accept json
// @Produce json
// @Param page query int false "Page number, default is 1"
// @Param size query int false "Page size, default is 10"
// @Param request_ids query []string false "List of request IDs to filter"
// @Param status query string false "Status filter (e.g., PENDING, SUCCESS, PARTIAL, EXPIRED, FAILED)"
// @Success 200 {object} dto.PaginationDTOResponse "Successful retrieval of token transfer histories"
// @Failure 400 {object} utils.GeneralError "Invalid parameters"
// @Failure 500 {object} utils.GeneralError "Internal server error"
// @Router /api/v1/payment-orders/histories [get]
func (h *paymentOrderHandler) GetPaymentOrderHistories(ctx *gin.Context) {
	// Set default values for page and size if they are not provided
	page := ctx.DefaultQuery("page", "1")
	size := ctx.DefaultQuery("size", "10")

	// Parse page and size into integers
	pageInt, err := strconv.Atoi(page)
	if err != nil || pageInt < 1 {
		log.LG.Errorf("Invalid page number: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid page number"})
		return
	}

	sizeInt, err := strconv.Atoi(size)
	if err != nil || sizeInt < 1 {
		log.LG.Errorf("Invalid size: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid size"})
		return
	}

	// Extract and parse request IDs from query string
	requestIDsStr := ctx.Query("request_ids")
	var requestIDs []string
	if requestIDsStr != "" {
		requestIDs = strings.Split(requestIDsStr, ",")
	}

	// Parse and validate `status` parameter if provided
	statusParam := ctx.Query("status")
	var status *string
	if statusParam != "" {
		status = &statusParam
	}

	// Call the use case to get payment order histories
	response, err := h.ucase.GetPaymentOrderHistories(ctx, requestIDs, status, pageInt, sizeInt)
	if err != nil {
		log.LG.Errorf("Failed to retrieve payment order histories: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to retrieve payment order histories",
			"details": err.Error(),
		})
		return
	}

	// Return the response as a JSON response
	ctx.JSON(http.StatusOK, response)
}
