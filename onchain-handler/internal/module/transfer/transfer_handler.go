package transfer

import (
	"net/http"

	"github.com/ethereum/go-ethereum/common"
	"github.com/gin-gonic/gin"

	"github.com/genefriendway/onchain-handler/internal/dto"
	"github.com/genefriendway/onchain-handler/internal/interfaces"
	"github.com/genefriendway/onchain-handler/internal/utils/log"
)

type TransferHandler struct {
	UCase interfaces.TransferUCase
}

// NewRewardHandler initializes the RewardHandler
func NewTransferHandler(ucase interfaces.TransferUCase) *TransferHandler {
	return &TransferHandler{
		UCase: ucase,
	}
}

// Transfer handles the distribution of tokens to recipients.
// @Summary Distribute tokens to recipients
// @Description This endpoint allows the distribution of tokens to multiple recipients. It accepts a list of transfer requests, validates the payload, and processes the token transfers based on the transaction type.
// @Tags transfer
// @Accept json
// @Produce json
// @Param payload body []dto.TransferTokenPayloadDTO true "List of transfer requests. Each request must include recipient address and transaction type."
// @Success 200 {object} map[string]bool "Success response: {\"success\": true}"
// @Failure 400 {object} util.GeneralError "Invalid payload or invalid recipient address/transaction type"
// @Failure 500 {object} util.GeneralError "Internal server error, failed to distribute tokens"
// @Router /api/v1/transfer [post]
func (h *TransferHandler) Transfer(ctx *gin.Context) {
	var req []dto.TransferTokenPayloadDTO

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

		// Check if the transaction type is valid
		if payload.TxType != "PURCHASE" && payload.TxType != "COMMISSION" {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error":   "Invalid tx_type",
				"details": "tx_type must be either PURCHASE or COMMISSION",
			})
			return
		}
	}

	// Proceed to distribute tokens if all checks pass
	if err := h.UCase.DistributeTokens(ctx, req); err != nil {
		log.LG.Errorf("Failed to distribute tokens: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to distribute tokens",
			"details": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"success": true})
}
