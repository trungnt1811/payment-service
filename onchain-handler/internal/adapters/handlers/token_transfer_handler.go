package handlers

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/gin-gonic/gin"

	"github.com/genefriendway/onchain-handler/conf"
	"github.com/genefriendway/onchain-handler/constants"
	"github.com/genefriendway/onchain-handler/internal/dto"
	"github.com/genefriendway/onchain-handler/internal/interfaces"
	httpresponse "github.com/genefriendway/onchain-handler/pkg/http/response"
	"github.com/genefriendway/onchain-handler/pkg/logger"
	"github.com/genefriendway/onchain-handler/pkg/utils"
)

type tokenTransferHandler struct {
	ucase  interfaces.TokenTransferUCase
	config *conf.Configuration
}

// NewRewardHandler initializes the RewardHandler
func NewTokenTransferHandler(ucase interfaces.TokenTransferUCase, config *conf.Configuration) *tokenTransferHandler {
	return &tokenTransferHandler{
		ucase:  ucase,
		config: config,
	}
}

// Transfer handles the distribution of tokens to recipients.
// @Summary Distribute tokens to recipients
// @Description This endpoint allows the distribution of tokens to multiple recipients. It accepts a list of transfer requests, validates the payload, and processes the token transfers based on the transaction type.
// @Tags token-transfer
// @Accept json
// @Produce json
// @Param payload body []dto.TokenTransferPayloadDTO true "List of transfer requests. Each request must include request id, from address, to address, amount, symbol (USDT) and network (AVAX C-Chain)."
// @Success 200 {object} []dto.TokenTransferResultDTOResponse "Successful distribution of tokens"
// @Failure 400 {object} response.GeneralError "Invalid payload or invalid recipient address/transaction type"
// @Failure 500 {object} response.GeneralError "Internal server error, failed to distribute tokens"
// @Router /api/v1/token-transfer [post]
func (h *tokenTransferHandler) Transfer(ctx *gin.Context) {
	var req []dto.TokenTransferPayloadDTO

	// Parse and validate the request payload
	if err := ctx.ShouldBindJSON(&req); err != nil {
		logger.GetLogger().Errorf("Invalid payload: %v", err)
		httpresponse.Error(ctx, http.StatusBadRequest, "Failed to distribute tokens, invalid payload", fmt.Errorf("invalid payload: %v", err))
		return
	}

	for index, payload := range req {
		// Check if the sender address is a valid Ethereum address
		if !common.IsHexAddress(payload.FromAddress) {
			fromAddress, err := h.config.GetPoolAddress(payload.FromAddress)
			if err != nil {
				httpresponse.Error(ctx, http.StatusBadRequest, "Failed to distribute tokens, invalid sender address", fmt.Errorf("invalid sender address: %s", payload.FromAddress))
				return
			}
			req[index].FromAddress = fromAddress
		}

		// Check if the recipient address is a valid Ethereum address
		if !common.IsHexAddress(payload.ToAddress) {
			toAddress, err := h.config.GetPoolAddress(payload.ToAddress)
			if err != nil {
				httpresponse.Error(ctx, http.StatusBadRequest, "Failed to distribute tokens, invalid recipient address", fmt.Errorf("invalid recipient address: %s", payload.ToAddress))
				return
			}
			req[index].ToAddress = toAddress
		}

		// Check if payload.Symbol is USDT
		if payload.Symbol != constants.USDT {
			httpresponse.Error(ctx, http.StatusBadRequest, "Failed to distribute tokens, invalid token symbol", fmt.Errorf("invalid token symbol: %s. Token symbol must be USDT", payload.Symbol))
			return
		}

		// Check if payload.Network is Avax C-Chain
		if payload.Network != string(constants.AvaxCChain) {
			httpresponse.Error(ctx, http.StatusBadRequest, "Failed to distribute tokens, invalid network", fmt.Errorf("invalid network: %s. Network must be Avax C-Chain", payload.Network))
			return
		}
	}

	// Proceed to distribute tokens if all checks pass
	transferResults, err := h.ucase.TransferTokens(ctx, req)
	if err != nil {
		logger.GetLogger().Errorf("Failed to distribute tokens: %v", err)
		httpresponse.Error(ctx, http.StatusInternalServerError, "Failed to distribute tokens", err)
		return
	}

	ctx.JSON(http.StatusOK, transferResults)
}

// GetTokenTransferHistories retrieves the token transfer histories.
// @Summary Get list of token transfer histories
// @Description This endpoint fetches a paginated list of token transfer histories filtered by request IDs and time range.
// @Tags token-transfer
// @Accept json
// @Produce json
// @Param page query int false "Page number, default is 1"
// @Param size query int false "Page size, default is 10"
// @Param request_ids query []string false "List of request IDs to filter"
// @Param start_time query string false "Start time in RFC3339 format to filter (e.g., 2024-01-01T00:00:00Z)"
// @Param end_time query string false "End time in RFC3339 format to filter (e.g., 2024-02-01T00:00:00Z)"
// @Param sort query string false "Sorting parameter in the format `field_direction` (e.g., id_asc, created_at_desc)"
// @Success 200 {object} dto.PaginationDTOResponse "Successful retrieval of token transfer histories"
// @Failure 400 {object} response.GeneralError "Invalid parameters"
// @Failure 500 {object} response.GeneralError "Internal server error"
// @Router /api/v1/token-transfers [get]
func (h *tokenTransferHandler) GetTokenTransferHistories(ctx *gin.Context) {
	page, size, err := utils.ParsePaginationParams(ctx)
	if err != nil {
		logger.GetLogger().Errorf("Invalid pagination parameters: %v", err)
		httpresponse.Error(ctx, http.StatusBadRequest, "Failed to retrieve token transfer histories, invalid pagination parameters", err)
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

	// Parse and validate the start_time query parameter
	startTimeStr := ctx.Query("start_time")
	startTime, err := time.Parse(time.RFC3339, startTimeStr)
	if startTimeStr != "" && err != nil {
		logger.GetLogger().Errorf("Invalid start time: %v", err)
		httpresponse.Error(ctx, http.StatusBadRequest, "Failed to retrieve token transfer histories, invalid start time", fmt.Errorf("invalid start time. Must be in RFC3339 format"))
		return
	}

	// Parse and validate the end_time query parameter
	endTimeStr := ctx.Query("end_time")
	endTime, err := time.Parse(time.RFC3339, endTimeStr)
	if endTimeStr != "" && err != nil {
		logger.GetLogger().Errorf("Invalid end time: %v", err)
		httpresponse.Error(ctx, http.StatusBadRequest, "Failed to retrieve token transfer histories, invalid end time", fmt.Errorf("invalid end time. Must be in RFC3339 format"))
		return
	}

	// Validate that start time is before end time
	if !startTime.IsZero() && !endTime.IsZero() && startTime.After(endTime) {
		httpresponse.Error(ctx, http.StatusBadRequest, "Failed to retrieve token transfer histories, start time must be before end time", fmt.Errorf("start time must be before end time"))
		return
	}

	// Parse and validate the `sort` parameter
	sort := ctx.Query("sort")
	orderBy, orderDirection, err := utils.ParseSortParameter(sort)
	if err != nil {
		logger.GetLogger().Errorf("Invalid sort parameter: %v", err)
		httpresponse.Error(ctx, http.StatusBadRequest, "Invalid sort parameter", err)
		return
	}

	// Fetch token transfer histories using the use case, passing request IDs, time range, page, and size
	response, err := h.ucase.GetTokenTransferHistories(
		ctx, requestIDs, startTime, endTime, orderBy, orderDirection, page, size,
	)
	if err != nil {
		logger.GetLogger().Errorf("Failed to retrieve token transfer histories: %v", err)
		httpresponse.Error(ctx, http.StatusInternalServerError, "Failed to retrieve token transfer histories", err)
		return
	}

	// Return the response as a JSON response
	ctx.JSON(http.StatusOK, response)
}
