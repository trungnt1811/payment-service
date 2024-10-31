package token_transfer

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/gin-gonic/gin"

	"github.com/genefriendway/onchain-handler/conf"
	"github.com/genefriendway/onchain-handler/constants"
	"github.com/genefriendway/onchain-handler/internal/dto"
	"github.com/genefriendway/onchain-handler/internal/interfaces"
	util "github.com/genefriendway/onchain-handler/internal/utils"
	"github.com/genefriendway/onchain-handler/log"
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
// @Param payload body []dto.TokenTransferPayloadDTO true "List of transfer requests. Each request must include recipient address and transaction type."
// @Success 200 {object} map[string]interface{} "Success response: {\"success\": true, \"results\": [{\"request_id\": \"requestID1\", \"status\": true, \"error_message\": \"\"}, {\"request_id\": \"requestID2\", \"status\": false, \"error_message\": \"Failed: some error message\"}]}"
// @Failure 400 {object} utils.GeneralError "Invalid payload or invalid recipient address/transaction type"
// @Failure 500 {object} utils.GeneralError "Internal server error, failed to distribute tokens"
// @Router /api/v1/token-transfer [post]
func (h *tokenTransferHandler) Transfer(ctx *gin.Context) {
	var req []dto.TokenTransferPayloadDTO

	// Parse and validate the request payload
	if err := ctx.ShouldBindJSON(&req); err != nil {
		log.LG.Errorf("Invalid payload: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid payload",
			"details": err.Error(),
		})
		return
	}

	for index, payload := range req {
		// Check if the sender address is a valid Ethereum address
		if !common.IsHexAddress(payload.FromAddress) {
			fromAddress, err := h.config.GetPoolAddress(payload.FromAddress)
			if err != nil {
				validPools := getValidPools()
				ctx.JSON(http.StatusBadRequest, gin.H{
					"error":   "Invalid recipient address: " + payload.ToAddress,
					"details": "SenderAddress must be a valid Ethereum address or must be one of the recognized pools: " + validPools,
				})
				return
			}
			req[index].FromAddress = fromAddress
		}

		// Check if the recipient address is a valid Ethereum address
		if !common.IsHexAddress(payload.ToAddress) {
			toAddress, err := h.config.GetPoolAddress(payload.ToAddress)
			if err != nil {
				validPools := getValidPools()
				ctx.JSON(http.StatusBadRequest, gin.H{
					"error":   "Invalid recipient address: " + payload.ToAddress,
					"details": "RecipientAddress must be a valid Ethereum address or must be one of the recognized pools: " + validPools,
				})
				return
			}
			req[index].ToAddress = toAddress
		}
		// Check if payload.Symbol is USDT or LP
		if payload.Symbol != constants.USDT && payload.Symbol != constants.LP {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error":   "Invalid token symbol",
				"details": "Token symbol must be USDT or LP",
			})
			return
		}
	}

	// Proceed to distribute tokens if all checks pass
	transferResults, err := h.ucase.TransferTokens(ctx, req)
	if err != nil {
		log.LG.Errorf("Failed to distribute tokens: %v", err)
		util.RespondError(ctx, http.StatusInternalServerError, "Failed to distribute tokens", err)
		return
	}

	// Return the result of each transaction (success or failure)
	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"results": transferResults,
	})
}

// getValidPools returns a comma-separated string of all valid pool names
func getValidPools() string {
	return fmt.Sprintf("%s, %s, %s, %s, %s",
		constants.LPCommunity,
		constants.LPStaking,
		constants.LPRevenue,
		constants.LPTreasury,
		constants.USDTTreasury,
	)
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
// @Param start_time query string false "Start time in RFC3339 format to filter example("2024-01-01T00:00:00Z")"
// @Param end_time query string false "End time in RFC3339 format to filter example("2024-02-01T00:00:00Z")"
// @Success 200 {object} dto.PaginationDTOResponse "Successful retrieval of token transfer histories"
// @Failure 400 {object} utils.GeneralError "Invalid parameters"
// @Failure 500 {object} utils.GeneralError "Internal server error"
// @Router /api/v1/token-transfer/histories [get]
func (h *tokenTransferHandler) GetTokenTransferHistories(ctx *gin.Context) {
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

	// Parse and validate the start_time query parameter
	startTimeStr := ctx.Query("start_time")
	startTime, err := time.Parse(time.RFC3339, startTimeStr)
	if startTimeStr != "" && err != nil {
		log.LG.Errorf("Invalid start time: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start time. Must be in RFC3339 format."})
		return
	}

	// Parse and validate the end_time query parameter
	endTimeStr := ctx.Query("end_time")
	endTime, err := time.Parse(time.RFC3339, endTimeStr)
	if endTimeStr != "" && err != nil {
		log.LG.Errorf("Invalid end time: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end time. Must be in RFC3339 format."})
		return
	}

	// Validate that start time is before end time
	if !startTime.IsZero() && !endTime.IsZero() && startTime.After(endTime) {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Start time must be before end time."})
		return
	}

	// Fetch token transfer histories using the use case, passing request IDs, time range, page, and size
	response, err := h.ucase.GetTokenTransferHistories(ctx, requestIDs, startTime, endTime, pageInt, sizeInt)
	if err != nil {
		log.LG.Errorf("Failed to retrieve token transfer histories: %v", err)
		util.RespondError(ctx, http.StatusInternalServerError, "Failed to retrieve token transfer histories", err)
		return
	}

	// Return the response as a JSON response
	ctx.JSON(http.StatusOK, response)
}
