package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/genefriendway/onchain-handler/internal/interfaces"
	httpresponse "github.com/genefriendway/onchain-handler/pkg/http/response"
	"github.com/genefriendway/onchain-handler/pkg/logger"
	"github.com/genefriendway/onchain-handler/pkg/utils"
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
// @Tags payment-wallet
// @Accept json
// @Produce json
// @Param address path string true "Address"
// @Success 200 {object} dto.PaymentWalletDTO
// @Failure 400 {object} response.GeneralError "Invalid address"
// @Failure 500 {object} response.GeneralError "Internal server error"
// @Router /api/v1/payment-wallet/{address} [get]
func (h *paymentWalletHandler) GetPaymentWalletByAddress(ctx *gin.Context) {
	address := ctx.Param("address")
	if !utils.IsValidEthAddress(address) {
		logger.GetLogger().Errorf("Invalid order address: %v", address)
		httpresponse.Error(ctx, http.StatusBadRequest, "Failed to retrieve payment wallet, invalid address", fmt.Errorf("invalid address: %v", address))
		return
	}

	wallet, err := h.ucase.GetPaymentWalletByAddress(ctx, address)
	if err != nil {
		httpresponse.Error(ctx, http.StatusInternalServerError, "Failed to retrieve payment wallet by address", err)
		return
	}
	ctx.JSON(http.StatusOK, wallet)
}

// GetPaymentWalletsWithBalances retrieves all payment wallets with their balances.
// @Summary Retrieves all payment wallets with balances.
// @Description Retrieves all payment wallets with balances grouped by network and token.
// @Tags payment-wallet
// @Accept json
// @Produce json
// @Success 200 {array} dto.PaymentWalletBalanceDTO
// @Failure 500 {object} response.GeneralError "Internal server error"
// @Router /api/v1/payment-wallets/balances [get]
func (h *paymentWalletHandler) GetPaymentWalletsWithBalances(ctx *gin.Context) {
	wallets, err := h.ucase.GetPaymentWalletsWithBalances(ctx, false, nil)
	if err != nil {
		logger.GetLogger().Errorf("Failed to retrieve payment wallets with balances: %v", err)
		httpresponse.Error(ctx, http.StatusInternalServerError, "Failed to retrieve payment wallets with balances", err)
		return
	}

	ctx.JSON(http.StatusOK, wallets)
}
