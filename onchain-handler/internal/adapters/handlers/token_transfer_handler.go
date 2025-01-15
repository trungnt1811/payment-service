package handlers

import (
	"fmt"
	"net/http"

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
	paymentWalletUCase interfaces.PaymentWalletUCase
	tokenTransferUCase interfaces.TokenTransferUCase
	config             *conf.Configuration
}

// NewRewardHandler initializes the RewardHandler
func NewTokenTransferHandler(
	paymentWalletUCase interfaces.PaymentWalletUCase,
	tokenTransferUCase interfaces.TokenTransferUCase,
	config *conf.Configuration,
) *tokenTransferHandler {
	return &tokenTransferHandler{
		paymentWalletUCase: paymentWalletUCase,
		tokenTransferUCase: tokenTransferUCase,
		config:             config,
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
		if payload.Network != constants.AvaxCChain.String() {
			httpresponse.Error(ctx, http.StatusBadRequest, "Failed to distribute tokens, invalid network", fmt.Errorf("invalid network: %s. Network must be Avax C-Chain", payload.Network))
			return
		}
	}

	// Proceed to distribute tokens if all checks pass
	transferResults, err := h.tokenTransferUCase.TransferTokens(ctx, req)
	if err != nil {
		logger.GetLogger().Errorf("Failed to distribute tokens: %v", err)
		httpresponse.Error(ctx, http.StatusInternalServerError, "Failed to distribute tokens", err)
		return
	}

	ctx.JSON(http.StatusOK, transferResults)
}

// GetTokenTransferHistories retrieves the token transfer histories.
// @Summary Get list of token transfer histories
// @Description This endpoint fetches a paginated list of token transfer histories filtered by time range and addresses.
// @Tags token-transfer
// @Accept json
// @Produce json
// @Param page query int false "Page number, default is 1"
// @Param size query int false "Page size, default is 10"
// @Param start_time query int false "Start time in UNIX timestamp format"
// @Param end_time query int false "End time in UNIX timestamp format"
// @Param from_address query string false "Filter by sender address"
// @Param to_address query string false "Filter by recipient address"
// @Param sort query string false "Sorting parameter in the format `field_direction` (e.g., id_asc, created_at_desc)"
// @Success 200 {object} dto.PaginationDTOResponse "Successful retrieval of token transfer histories"
// @Failure 400 {object} response.GeneralError "Invalid parameters"
// @Failure 500 {object} response.GeneralError "Internal server error"
// @Router /api/v1/token-transfers [get]
func (h *tokenTransferHandler) GetTokenTransferHistories(ctx *gin.Context) {
	// Parse pagination parameters
	page, size, err := utils.ParsePaginationParams(ctx)
	if err != nil {
		logger.GetLogger().Errorf("Invalid pagination parameters: %v", err)
		httpresponse.Error(ctx, http.StatusBadRequest, "Failed to retrieve token transfer histories, invalid pagination parameters", err)
		return
	}

	// Parse and validate start_time parameter
	startTime, err := utils.ParseOptionalUnixTimestamp(ctx.Query("start_time"))
	if err != nil {
		logger.GetLogger().Errorf("Invalid start_time: %v", err)
		httpresponse.Error(ctx, http.StatusBadRequest, "Invalid start_time. Provide a valid UNIX timestamp.", err)
		return
	}

	// Parse and validate end_time parameter
	endTime, err := utils.ParseOptionalUnixTimestamp(ctx.Query("end_time"))
	if err != nil {
		logger.GetLogger().Errorf("Invalid end_time: %v", err)
		httpresponse.Error(ctx, http.StatusBadRequest, "Invalid end_time. Provide a valid UNIX timestamp.", err)
		return
	}

	// Validate that start time is before end time
	if startTime != nil && endTime != nil && startTime.After(*endTime) {
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

	// Parse optional from_address and to_address query parameters
	var fromAddress *string
	if addr := ctx.Query("from_address"); addr != "" {
		fromAddress = &addr
	}

	var toAddress *string
	if addr := ctx.Query("to_address"); addr != "" {
		toAddress = &addr
	}

	// Call the use case to fetch token transfer histories
	response, err := h.tokenTransferUCase.GetTokenTransferHistories(
		ctx, startTime, endTime, orderBy, orderDirection, page, size, fromAddress, toAddress,
	)
	if err != nil {
		logger.GetLogger().Errorf("Failed to retrieve token transfer histories: %v", err)
		httpresponse.Error(ctx, http.StatusInternalServerError, "Failed to retrieve token transfer histories", err)
		return
	}

	// Return the response
	ctx.JSON(http.StatusOK, response)
}

// GetWithdrawHistories retrieves the withdraw histories.
// @Summary Get list of withdraw histories
// @Description Fetches a paginated list of withdraw histories filtered by time range, sender, and recipient addresses.
// @Tags withdraw
// @Accept json
// @Produce json
// @Param page query int false "Page number, default is 1"
// @Param size query int false "Page size, default is 10"
// @Param start_time query int false "Start time in UNIX timestamp format"
// @Param end_time query int false "End time in UNIX timestamp format"
// @Param to_address query string false "Filter by recipient address"
// @Param sort query string false "Sorting parameter in the format `field_direction` (e.g., id_asc, created_at_desc)"
// @Success 200 {object} dto.PaginationDTOResponse "Successful retrieval of withdraw histories"
// @Failure 400 {object} response.GeneralError "Invalid parameters"
// @Failure 500 {object} response.GeneralError "Internal server error"
// @Router /api/v1/withdraws [get]
func (h *tokenTransferHandler) GetWithdrawHistories(ctx *gin.Context) {
	// Parse pagination parameters
	page, size, err := utils.ParsePaginationParams(ctx)
	if err != nil {
		logger.GetLogger().Errorf("Invalid pagination parameters: %v", err)
		httpresponse.Error(ctx, http.StatusBadRequest, "Invalid pagination parameters", err)
		return
	}

	// Parse and validate start_time parameter
	startTime, err := utils.ParseOptionalUnixTimestamp(ctx.Query("start_time"))
	if err != nil {
		logger.GetLogger().Errorf("Invalid start_time: %v", err)
		httpresponse.Error(ctx, http.StatusBadRequest, "Invalid start_time. Provide a valid UNIX timestamp.", err)
		return
	}

	// Parse and validate end_time parameter
	endTime, err := utils.ParseOptionalUnixTimestamp(ctx.Query("end_time"))
	if err != nil {
		logger.GetLogger().Errorf("Invalid end_time: %v", err)
		httpresponse.Error(ctx, http.StatusBadRequest, "Invalid end_time. Provide a valid UNIX timestamp.", err)
		return
	}

	// Validate that start time is before end time
	if startTime != nil && endTime != nil && startTime.After(*endTime) {
		httpresponse.Error(ctx, http.StatusBadRequest, "start_time must be before end_time", nil)
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

	// Parse optional to_address query parameter
	var toAddress *string
	if addr := ctx.Query("to_address"); addr != "" {
		toAddress = &addr
	}

	// Get the receiving wallet address
	receivingAddress, err := h.paymentWalletUCase.GetReceivingWalletAddress(
		ctx, h.config.Wallet.Mnemonic, h.config.Wallet.Passphrase, h.config.Wallet.Salt,
	)
	if err != nil {
		httpresponse.Error(ctx, http.StatusInternalServerError, "Failed to get receiving wallet address", err)
		return
	}

	// Call the use case to fetch withdraw histories
	response, err := h.tokenTransferUCase.GetTokenTransferHistories(
		ctx, startTime, endTime, orderBy, orderDirection, page, size, &receivingAddress, toAddress,
	)
	if err != nil {
		logger.GetLogger().Errorf("Failed to retrieve withdraw histories: %v", err)
		httpresponse.Error(ctx, http.StatusInternalServerError, "Failed to retrieve withdraw histories", err)
		return
	}

	// Call the use case to calculate total token amount
	totalTokenAmount, err := h.tokenTransferUCase.GetTotalTokenAmount(ctx, startTime, endTime, &receivingAddress, toAddress)
	if err != nil {
		httpresponse.Error(ctx, http.StatusInternalServerError, "Failed to calculate total token amount", err)
		return
	}

	// Set the total token amount in the response
	response.TotalTokenAmount = totalTokenAmount

	// Return the response
	ctx.JSON(http.StatusOK, response)
}
