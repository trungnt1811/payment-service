package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"gorm.io/gorm"

	"github.com/gin-gonic/gin"

	"github.com/genefriendway/onchain-handler/conf"
	"github.com/genefriendway/onchain-handler/constants"
	"github.com/genefriendway/onchain-handler/internal/domain/dto"
	"github.com/genefriendway/onchain-handler/internal/interfaces"
	"github.com/genefriendway/onchain-handler/pkg/database/postgresql"
	httpresponse "github.com/genefriendway/onchain-handler/pkg/http"
	"github.com/genefriendway/onchain-handler/pkg/logger"
	"github.com/genefriendway/onchain-handler/pkg/utils"
)

type paymentOrderHandler struct {
	ucase interfaces.PaymentOrderUCase
}

func NewPaymentOrderHandler(
	ucase interfaces.PaymentOrderUCase,
) *paymentOrderHandler {
	return &paymentOrderHandler{
		ucase: ucase,
	}
}

// CreateOrders creates payment orders based on the input payload.
// @Summary Create payment orders
// @Description This endpoint allows creating payment orders for users.
// @Tags payment-order
// @Accept json
// @Produce json
// @Param Vendor-Id header string true "Vendor ID for authentication"
// @Param payload body []dto.PaymentOrderPayloadDTO true "List of payment orders. Each order must include request id, amount, symbol (USDT) and network (AVAX C-Chain or BSC)."
// @Success 201 {object} map[string]interface{} "Success created: {\"success\": true, \"data\": []dto.CreatedPaymentOrderDTO}"
// @Failure 400 {object} http.GeneralError "Invalid payload"
// @Failure 412 {object} http.GeneralError "Duplicate key value"
// @Failure 500 {object} http.GeneralError "Internal server error"
// @Router /api/v1/payment-orders [post]
func (h *paymentOrderHandler) CreateOrders(ctx *gin.Context) {
	var req []dto.PaymentOrderPayloadDTO

	// Get the Vendor-Id from the header
	vendorID := ctx.GetHeader("Vendor-Id")

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
		if err := utils.ValidateNetworkType(order.Network); err != nil {
			logger.GetLogger().Errorf("Validation failed for network alias %s: %v", order.Network, err)
			httpresponse.Error(ctx, http.StatusBadRequest, fmt.Sprintf("Failed to create payment orders, unsupported network: %s", order.Network), err)
			return
		}
	}

	// Call the use case to create the payment orders
	response, err := h.ucase.CreatePaymentOrders(ctx, req, vendorID, conf.GetExpiredOrderTime())
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

// GetPaymentOrders retrieves payment orders optionally filtered by status, from_address, network, and sorted using the sort query parameter.
// @Summary Retrieve payment orders
// @Description This endpoint retrieves payment orders based on optional filters such as status, from_address, network, and sorting options.
// @Tags payment-order
// @Accept json
// @Produce json
// @Param Vendor-Id header string true "Vendor ID for authentication"
// @Param page query int false "Page number, default is 1"
// @Param size query int false "Page size, default is 10"
// @Param request_ids query []string false "List of request IDs to filter (maximum 50)"
// @Param from_address query string false "Filter by sender's address (from_address)"
// @Param network query string false "Filter by network (e.g., BSC, AVAX C-Chain)"
// @Param status query string false "Status filter (e.g., PENDING, PROCESSING, SUCCESS, PARTIAL, EXPIRED, FAILED)"
// @Param sort query string false "Sorting parameter in the format `field_direction` (e.g., id_asc, created_at_desc, succeeded_at_desc)"
// @Param start_time query int false "Start time in UNIX timestamp format to filter (e.g., 1704067200)"
// @Param end_time query int false "End time in UNIX timestamp format to filter (e.g., 1706745600)"
// @Param time_filter_field query string false "Field to filter time (e.g., created_at or succeeded_at)"
// @Success 200 {object} dto.PaginationDTOResponse "Successful retrieval of payment order histories"
// @Failure 400 {object} http.GeneralError "Invalid parameters"
// @Failure 500 {object} http.GeneralError "Internal server error"
// @Router /api/v1/payment-orders [get]
func (h *paymentOrderHandler) GetPaymentOrders(ctx *gin.Context) {
	// Get the Vendor-Id from the header
	vendorID := ctx.GetHeader("Vendor-Id")

	// Parse pagination parameters
	page, size, err := utils.ParsePaginationParams(ctx)
	if err != nil {
		logger.GetLogger().Errorf("Invalid pagination parameters: %v", err)
		httpresponse.Error(ctx, http.StatusBadRequest, "Failed to retrieve payment orders, invalid pagination parameters", err)
		return
	}

	// Extract and parse request IDs from query string
	requestIDsParam := ctx.Query("request_ids")
	requestIDsMap := make(map[string]struct{}) // To handle duplicates
	var requestIDs []string

	if requestIDsParam != "" {
		for _, id := range strings.Split(requestIDsParam, ",") {
			trimmedID := strings.TrimSpace(id)
			if trimmedID != "" {
				if _, exists := requestIDsMap[trimmedID]; !exists {
					requestIDsMap[trimmedID] = struct{}{}
					requestIDs = append(requestIDs, trimmedID)
				}
			}
		}
	}

	// Check if the number of request IDs exceeds the limit
	if len(requestIDs) > 50 {
		logger.GetLogger().Errorf("Too many request IDs: received %d, maximum allowed is 50", len(requestIDs))
		httpresponse.Error(ctx, http.StatusBadRequest, "Too many request IDs. Maximum allowed is 50", nil)
		return
	}

	// Parse optional query parameters
	status := utils.ParseOptionalQuery(ctx.Query("status"))
	fromAddress := utils.ParseOptionalQuery(ctx.Query("from_address"))
	if fromAddress != nil && !utils.IsValidEthAddress(*fromAddress) {
		emptyResponse := dto.PaginationDTOResponse{
			Page:     page,
			Size:     size,
			NextPage: page,
			Data:     nil,
		}
		ctx.JSON(http.StatusOK, emptyResponse)
		return
	}
	network := utils.ParseOptionalQuery(ctx.Query("network"))
	if network != nil {
		if err := utils.ValidateNetworkType(*network); err != nil {
			logger.GetLogger().Errorf("Invalid network: %v", err)
			httpresponse.Error(ctx, http.StatusBadRequest, fmt.Sprintf("Unsupported network: %s", *network), err)
			return
		}
	}

	timeFilterField := ctx.Query("time_filter_field")
	// Default to created_at if time_filter_field is not provided
	if timeFilterField == "" {
		timeFilterField = "created_at"
	}
	// Validate time_filter_field
	if timeFilterField != "" && timeFilterField != "created_at" && timeFilterField != "succeeded_at" {
		logger.GetLogger().Errorf("Invalid time_filter_field: %s", timeFilterField)
		httpresponse.Error(ctx, http.StatusBadRequest, "Invalid time_filter_field. Use 'created_at' or 'succeeded_at'.", nil)
		return
	}

	// Parse and validate time parameters
	startTime, err := utils.ParseOptionalUnixTimestamp(ctx.Query("start_time"))
	if err != nil {
		logger.GetLogger().Errorf("Invalid start_time: %v", err)
		httpresponse.Error(ctx, http.StatusBadRequest, "Invalid start_time. Provide a valid UNIX timestamp.", err)
		return
	}
	endTime, err := utils.ParseOptionalUnixTimestamp(ctx.Query("end_time"))
	if err != nil {
		logger.GetLogger().Errorf("Invalid end_time: %v", err)
		httpresponse.Error(ctx, http.StatusBadRequest, "Invalid end_time. Provide a valid UNIX timestamp.", err)
		return
	}
	if startTime != nil && endTime != nil && startTime.After(*endTime) {
		httpresponse.Error(ctx, http.StatusBadRequest, "start_time must be earlier than end_time", nil)
		return
	}

	// Parse and validate sort parameter
	sort := ctx.Query("sort")
	orderBy, orderDirection, err := utils.ParseSortParameter(sort)
	if err != nil {
		logger.GetLogger().Errorf("Invalid sort parameter: %v", err)
		httpresponse.Error(ctx, http.StatusBadRequest, "Invalid sort parameter", err)
		return
	}

	// Call the use case to get payment orders
	response, err := h.ucase.GetPaymentOrders(
		ctx, vendorID, requestIDs, status, orderBy, fromAddress, network, orderDirection, startTime, endTime, &timeFilterField, page, size,
	)
	if err != nil {
		logger.GetLogger().Errorf("Failed to retrieve payment orders: %v", err)
		httpresponse.Error(ctx, http.StatusInternalServerError, "Failed to retrieve payment orders", err)
		return
	}

	// Return the response
	ctx.JSON(http.StatusOK, response)
}

// GetPaymentOrderByRequestID retrieves a payment order by its request ID.
// @Summary Retrieve payment order by request ID
// @Description This endpoint retrieves a payment order by its request ID, which can contain special characters.
// @Tags payment-order
// @Accept json
// @Produce json
// @Param request_id path string true "Payment order request ID"
// @Success 200 {object} dto.PaymentOrderDTOResponse "Successful retrieval of payment order"
// @Failure 400 {object} http.GeneralError "Invalid request ID"
// @Failure 404 {object} http.GeneralError "Payment order not found"
// @Failure 500 {object} http.GeneralError "Internal server error"
// @Router /api/v1/payment-order/{request_id} [get]
func (h *paymentOrderHandler) GetPaymentOrderByRequestID(ctx *gin.Context) {
	// Extract request ID directly as a string
	requestID := ctx.Param("request_id")
	if requestID == "" {
		logger.GetLogger().Error("Request ID is empty")
		httpresponse.Error(ctx, http.StatusBadRequest, "Failed to retrieve payment order, request ID cannot be empty", nil)
		return
	}

	// Delegate to the use case layer
	response, err := h.ucase.GetPaymentOrderByRequestID(ctx, requestID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.GetLogger().Warnf("Payment order not found for request ID %s", requestID)
			httpresponse.Error(ctx, http.StatusNotFound, "Payment order not found", nil)
			return
		}

		logger.GetLogger().Errorf("Failed to retrieve payment order for request ID %s: %v", requestID, err)
		httpresponse.Error(ctx, http.StatusInternalServerError, "Failed to retrieve payment order", err)
		return
	}

	// Return the payment order in JSON format
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
// @Failure 400 {object} http.GeneralError "Invalid payload"
// @Failure 400 {object} http.GeneralError "Unsupported network"
// @Failure 500 {object} http.GeneralError "Internal server error"
// @Router /api/v1/payment-order/network [put]
func (h *paymentOrderHandler) UpdatePaymentOrderNetwork(ctx *gin.Context) {
	var req dto.PaymentOrderNetworkPayloadDTO

	// Parse and validate the request payload
	if err := ctx.ShouldBindJSON(&req); err != nil {
		logger.GetLogger().Errorf("Invalid payload: %v", err)
		httpresponse.Error(ctx, http.StatusBadRequest, "Failed to update payment order network, invalid payload", err)
		return
	}

	if err := utils.ValidateNetworkType(req.Network); err != nil {
		logger.GetLogger().Errorf("Unsupported network: %s", req.Network)
		httpresponse.Error(ctx, http.StatusBadRequest, fmt.Sprintf("Failed to update payment order network, unsupported network: %s", err))
		return
	}

	// Determine the network type
	var network constants.NetworkType
	if req.Network == constants.AvaxCChain.String() {
		network = constants.AvaxCChain
	} else {
		network = constants.Bsc
	}

	// Call the use case to update the payment order network
	if err := h.ucase.UpdateOrderNetwork(ctx, req.RequestID, network); err != nil {
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
