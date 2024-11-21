package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/genefriendway/onchain-handler/conf"
	"github.com/genefriendway/onchain-handler/constants"
	"github.com/genefriendway/onchain-handler/internal/dto"
	"github.com/genefriendway/onchain-handler/internal/interfaces"
	"github.com/genefriendway/onchain-handler/pkg/database/postgresql"
	httpresponse "github.com/genefriendway/onchain-handler/pkg/http/response"
	"github.com/genefriendway/onchain-handler/pkg/logger"
	"github.com/genefriendway/onchain-handler/pkg/utils"
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
// @Param payload body []dto.PaymentOrderPayloadDTO true "List of payment orders. Each order must include request id, amount, symbol (USDT) and network (AVAX C-Chain or BSC)."
// @Success 201 {object} map[string]interface{} "Success created: {\"success\": true, \"data\": []dto.CreatedPaymentOrderDTO}"
// @Failure 400 {object} response.GeneralError "Invalid payload"
// @Failure 412 {object} response.GeneralError "Duplicate key value"
// @Failure 500 {object} response.GeneralError "Internal server error"
// @Router /api/v1/payment-orders [post]
func (h *paymentOrderHandler) CreateOrders(ctx *gin.Context) {
	var req []dto.PaymentOrderPayloadDTO

	// Parse and validate the request payload
	if err := ctx.ShouldBindJSON(&req); err != nil {
		logger.GetLogger().Errorf("Invalid payload: %v", err)
		httpresponse.Error(ctx, http.StatusBadRequest, "Failed to create payment orders, invalid payload", err)
		return
	}

	// Validate each payment order
	for _, order := range req {
		if err := validatePaymentOrder(order); err != nil {
			logger.GetLogger().Errorf("Validation failed for request id %s: %v", order.RequestID, err)
			httpresponse.Error(ctx, http.StatusBadRequest, fmt.Sprintf("Failed to create payment orders, validation failed for request id: %s", order.RequestID), err)
			return
		}
		if err := validateNetworkType(order.Network); err != nil {
			logger.GetLogger().Errorf("Validation failed for request id %s: %v", order.RequestID, err)
			httpresponse.Error(ctx, http.StatusBadRequest, fmt.Sprintf("Failed to create payment orders, unsupported network: %s", order.RequestID), err)
			return
		}
	}

	// Call the use case to create the payment orders
	response, err := h.ucase.CreatePaymentOrders(ctx, req, h.config.GetExpiredOrderTime())
	if err != nil {
		logger.GetLogger().Errorf("Failed to create payment orders: %v", err)
		if postgresql.IsUniqueViolation(err) {
			httpresponse.Error(ctx, http.StatusPreconditionFailed, "Failed to create payment orders, duplicate key value violates unique constraint", err)
			return
		} else {
			httpresponse.Error(ctx, http.StatusInternalServerError, "Failed to create payment orders", err)
			return
		}
	}

	// Respond with success and response data
	ctx.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    response,
	})
}

// GetPaymentOrders retrieves payment orders optionally filters by status.
// @Summary Retrieve payment orders
// @Description This endpoint retrieves payment orders based on optional status filter.
// @Tags payment-order
// @Accept json
// @Produce json
// @Param page query int false "Page number, default is 1"
// @Param size query int false "Page size, default is 10"
// @Param status query string false "Status filter (e.g., PENDING, SUCCESS, PARTIAL, EXPIRED, FAILED)"
// @Success 200 {object} dto.PaginationDTOResponse "Successful retrieval of payment order histories"
// @Failure 400 {object} response.GeneralError "Invalid parameters"
// @Failure 500 {object} response.GeneralError "Internal server error"
// @Router /api/v1/payment-orders [get]
func (h *paymentOrderHandler) GetPaymentOrders(ctx *gin.Context) {
	// Parse pagination parameters
	page, size, err := utils.ParsePaginationParams(ctx)
	if err != nil {
		logger.GetLogger().Errorf("Invalid pagination parameters: %v", err)
		httpresponse.Error(ctx, http.StatusBadRequest, "Failed to retrieve payment orders, invalid pagination parameters", err)
		return
	}

	// Parse `status` query parameter
	status := parseOptionalQuery(ctx.Query("status"))

	// Call the use case to get payment orders
	response, err := h.ucase.GetPaymentOrders(ctx, status, page, size)
	if err != nil {
		logger.GetLogger().Errorf("Failed to retrieve payment orders: %v", err)
		httpresponse.Error(ctx, http.StatusInternalServerError, "Failed to retrieve payment orders", err)
		return
	}

	// Return the response
	ctx.JSON(http.StatusOK, response)
}

func parseOptionalQuery(param string) *string {
	if param == "" {
		return nil
	}
	return &param
}

// GetPaymentOrderByID retrieves a payment order by its ID.
// @Summary Retrieve payment order by ID
// @Description This endpoint retrieves a payment order by its ID.
// @Tags payment-order
// @Accept json
// @Produce json
// @Param id path int true "Payment order ID"
// @Success 200 {object} dto.PaymentOrderDTOResponse "Successful retrieval of payment order"
// @Failure 400 {object} response.GeneralError "Invalid order ID"
// @Failure 500 {object} response.GeneralError "Internal server error"
// @Router /api/v1/payment-order/{id} [get]
func (h *paymentOrderHandler) GetPaymentOrderByID(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		logger.GetLogger().Errorf("Invalid order ID: %v", err)
		httpresponse.Error(ctx, http.StatusBadRequest, "Failed to retrieve payment order, invalid order ID", err)
		return
	}

	response, err := h.ucase.GetPaymentOrderByID(ctx, id)
	if err != nil {
		logger.GetLogger().Errorf("Failed to retrieve payment order: %v", err)
		httpresponse.Error(ctx, http.StatusInternalServerError, "Failed to retrieve payment order", err)
		return
	}

	ctx.JSON(http.StatusOK, response)
}

// UpdatePaymentOrderNetwork updates the network of a payment order.
// @Summary Update payment order network
// @Description This endpoint allows updating the network of a payment order.
// @Tags payment-order
// @Accept json
// @Produce json
// @Param payload body dto.PaymentOrderNetworkPayloadDTO true "Payment order ID and network (AVAX C-Chain or BSC)."
// @Success 200 {object} map[string]interface{} "Success response: {\"success\": true}"
// @Failure 400 {object} response.GeneralError "Invalid payload"
// @Failure 400 {object} response.GeneralError "Unsupported network"
// @Failure 500 {object} response.GeneralError "Internal server error"
// @Router /api/v1/payment-order/network [put]
func (h *paymentOrderHandler) UpdatePaymentOrderNetwork(ctx *gin.Context) {
	var req dto.PaymentOrderNetworkPayloadDTO

	// Parse and validate the request payload
	if err := ctx.ShouldBindJSON(&req); err != nil {
		logger.GetLogger().Errorf("Invalid payload: %v", err)
		httpresponse.Error(ctx, http.StatusBadRequest, "Failed to update payment order network, invalid payload", err)
		return
	}

	if err := validateNetworkType(req.Network); err != nil {
		logger.GetLogger().Errorf("Unsupported network: %s", req.Network)
		httpresponse.Error(ctx, http.StatusBadRequest, fmt.Sprintf("Failed to update payment order network, unsupported network: %s", err))
		return
	}

	// Call the use case to update the payment order network
	if err := h.ucase.UpdatePaymentOrder(ctx, req.ID, nil, nil, nil, &req.Network); err != nil {
		logger.GetLogger().Errorf("Failed to update payment order network: %v", err)
		httpresponse.Error(ctx, http.StatusInternalServerError, "Failed to update payment order network", err)
		return
	}

	// Respond with success
	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
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

	return nil
}

func validateNetworkType(network string) error {
	// Validate network type is either BSC or AVAX C-Chain
	validNetworks := map[constants.NetworkType]bool{
		constants.Bsc:        true,
		constants.AvaxCChain: true,
	}
	networkType := constants.NetworkType(network)
	if !validNetworks[networkType] {
		return fmt.Errorf("invalid network type: %s, must be BSC or AVAX C-Chain", network)
	}

	return nil
}
