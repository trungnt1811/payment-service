package token_transfer

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/ethereum/go-ethereum/common"
	"github.com/gin-gonic/gin"

	"github.com/genefriendway/onchain-handler/constants"
	"github.com/genefriendway/onchain-handler/internal/dto"
	"github.com/genefriendway/onchain-handler/internal/interfaces"
	util "github.com/genefriendway/onchain-handler/internal/utils"
	"github.com/genefriendway/onchain-handler/internal/utils/log"
)

type TokenTransferHandler struct {
	UCase interfaces.TokenTransferUCase
}

// NewRewardHandler initializes the RewardHandler
func NewTokenTransferHandler(ucase interfaces.TokenTransferUCase) *TokenTransferHandler {
	return &TokenTransferHandler{
		UCase: ucase,
	}
}

// Transfer handles the distribution of tokens to recipients.
// @Summary Distribute tokens to recipients
// @Description This endpoint allows the distribution of tokens to multiple recipients. It accepts a list of transfer requests, validates the payload, and processes the token transfers based on the transaction type.
// @Tags token-transfer
// @Accept json
// @Produce json
// @Param payload body []dto.TokenTransferPayloadDTO true "List of transfer requests. Each request must include recipient address and transaction type."
// @Success 200 {object} map[string]bool "Success response: {\"success\": true}"
// @Failure 400 {object} util.GeneralError "Invalid payload or invalid recipient address/transaction type"
// @Failure 500 {object} util.GeneralError "Internal server error, failed to distribute tokens"
// @Router /api/v1/token-transfer [post]
func (h *TokenTransferHandler) Transfer(ctx *gin.Context) {
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

	for _, payload := range req {
		// Validate the sender using constants for recognized pool names
		if !isValidPool(payload.PoolName) {
			validPools := getValidPools()
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error":   "Invalid sender pool: " + payload.PoolName,
				"details": "FromAddress must be one of the recognized pools: " + validPools,
			})
			return
		}

		// Check if the recipient address is a valid Ethereum address
		if !common.IsHexAddress(payload.ToAddress) {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error":   "Invalid recipient address: " + payload.ToAddress,
				"details": "RecipientAddress must be a valid Ethereum address",
			})
			return
		}
	}

	// Proceed to distribute tokens if all checks pass
	if err := h.UCase.TransferTokens(ctx, req); err != nil {
		log.LG.Errorf("Failed to distribute tokens: %v", err)
		util.RespondError(ctx, http.StatusInternalServerError, "Failed to distribute tokens", err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"success": true})
}

// isValidPool checks if the given pool name is valid by comparing it with known constants.
func isValidPool(poolName string) bool {
	switch poolName {
	case constants.LPCommunity, constants.LPStaking, constants.LPRevenue, constants.LPTreasury, constants.USDTTreasury:
		return true
	default:
		return false
	}
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
// @Description This endpoint fetches a paginated list of token transfer histories with optional filters.
// @Tags token-transfer
// @Accept json
// @Produce json
// @Param page query int false "Page number, default is 1"
// @Param size query int false "Page size, default is 10"
// @Param pool_name query string false "Pool's name to filter (LP_Treasury, LP_Revenue, LP_Staking, LP_Community, USDT_Treasury)"
// @Param transaction_hash query string false "Transaction hash to filter"
// @Param to_address query string false "Recipient's address to filter"
// @Param symbol query string false "Token symbol to filter"
// @Success 200 {object} dto.TokenTransferHistoryDTOResponse "Successful retrieval of token transfer histories"
// @Failure 400 {object} util.GeneralError "Invalid parameters"
// @Failure 500 {object} util.GeneralError "Internal server error"
// @Security ApiKeyAuth
// @Router /api/v1/token-transfer/histories [get]
func (h *TokenTransferHandler) GetTokenTransferHistories(ctx *gin.Context) {
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

	// Extract filter parameters from query string
	transactionHash := ctx.Query("transaction_hash")
	poolName := ctx.Query("pool_name")
	toAddress := ctx.Query("to_address")
	symbol := ctx.Query("symbol")

	// Create filter DTO
	filters := dto.TokenTransferFilterDTO{
		TransactionHash: &transactionHash,
		PoolName:        &poolName,
		ToAddress:       &toAddress,
		Symbol:          &symbol,
	}

	// Fetch token transfer histories using the use case, passing filters, page, and size
	response, err := h.UCase.GetTokenTransferHistories(ctx, filters, pageInt, sizeInt)
	if err != nil {
		log.LG.Errorf("Failed to retrieve token transfer histories: %v", err)
		util.RespondError(ctx, http.StatusInternalServerError, "Failed to retrieve token transfer histories", err)
		return
	}

	// Return the response as a JSON response
	ctx.JSON(http.StatusOK, response)
}
