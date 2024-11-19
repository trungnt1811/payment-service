package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/genefriendway/onchain-handler/internal/interfaces"
	"github.com/genefriendway/onchain-handler/pkg/blockchain/utils"
	"github.com/genefriendway/onchain-handler/pkg/logger"
)

type paymentWalletHandler struct {
	ucase interfaces.PaymentWalletUCase
}

func NewPaymentWalletHandler(ucase interfaces.PaymentWalletUCase) *paymentWalletHandler {
	return &paymentWalletHandler{
		ucase: ucase,
	}
}

// GetPaymentWalletByAddress retrieves a payment wallet by its address.
// @Summary Retrieves a payment wallet by its address.
// @Description Retrieves a payment wallet by its address.
// @Tags payment_wallet
// @Accept json
// @Produce json
// @Param address path string true "Address"
// @Success 200 {object} dto.PaymentWalletDTO
// @Failure 400 {object} response.GeneralError "Invalid address"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/payment-wallet/{address} [get]
func (h *paymentWalletHandler) GetPaymentWalletByAddress(ctx *gin.Context) {
	address := ctx.Param("address")
	if !utils.IsValidEthAddress(address) {
		logger.GetLogger().Errorf("Invalid order address: %v", address)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid address"})
		return
	}

	wallet, err := h.ucase.GetPaymentWalletByAddress(ctx, address)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"data": wallet})
}
