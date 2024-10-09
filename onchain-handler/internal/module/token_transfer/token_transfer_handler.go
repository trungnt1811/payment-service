package transfer

import (
	"net/http"

	"github.com/ethereum/go-ethereum/common"
	"github.com/gin-gonic/gin"

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
// @Tags transfer
// @Accept json
// @Produce json
// @Param payload body []dto.TokenTransferPayloadDTO true "List of transfer requests. Each request must include recipient address and transaction type."
// @Success 200 {object} map[string]bool "Success response: {\"success\": true}"
// @Failure 400 {object} util.GeneralError "Invalid payload or invalid recipient address/transaction type"
// @Failure 500 {object} util.GeneralError "Internal server error, failed to distribute tokens"
// @Router /api/v1/transfer [post]
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
		// Check if the recipient address is a valid Ethereum address
		if !common.IsHexAddress(payload.RecipientAddress) {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error":   "Invalid recipient address: " + payload.RecipientAddress,
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
