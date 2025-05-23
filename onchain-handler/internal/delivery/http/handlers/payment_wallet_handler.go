package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/genefriendway/onchain-handler/conf"
	"github.com/genefriendway/onchain-handler/constants"
	"github.com/genefriendway/onchain-handler/internal/delivery/dto"
	ucasetypes "github.com/genefriendway/onchain-handler/internal/domain/ucases/types"
	httpresponse "github.com/genefriendway/onchain-handler/pkg/http"
	"github.com/genefriendway/onchain-handler/pkg/logger"
	"github.com/genefriendway/onchain-handler/pkg/utils"
)

type paymentWalletHandler struct {
	ucase  ucasetypes.PaymentWalletUCase
	config *conf.Configuration
}

func NewPaymentWalletHandler(ucase ucasetypes.PaymentWalletUCase, config *conf.Configuration) *paymentWalletHandler {
	return &paymentWalletHandler{
		ucase:  ucase,
		config: config,
	}
}

// GetPaymentWalletByAddress retrieves a payment wallet by its address.
// @Summary Retrieves a payment wallet by its address.
// @Description Retrieves a payment wallet by its address.
// @Tags payment-wallet
// @Accept json
// @Produce json
// @Param address path string true "Address"
// @Success 200 {object} dto.PaymentWalletBalanceDTO
// @Failure 400 {object} http.GeneralError "Invalid address"
// @Failure 500 {object} http.GeneralError "Internal server error"
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
// @Description Retrieves all payment wallets with balances grouped by network and token. Supports optional filtering by network.
// @Tags payment-wallet
// @Accept json
// @Produce json
// @Param page query int false "Page number, default is 1"
// @Param size query int false "Page size, default is 10"
// @Param network query string false "Filter by network (e.g., BSC, AVAX C-Chain)"
// @Success 200 {array} dto.PaginationDTOResponse
// @Failure 400 {object} http.GeneralError "Invalid network"
// @Failure 500 {object} http.GeneralError "Internal server error"
// @Router /api/v1/payment-wallets/balances [get]
func (h *paymentWalletHandler) GetPaymentWalletsWithBalances(ctx *gin.Context) {
	// Parse pagination parameters
	page, size, err := utils.ParsePaginationParams(ctx)
	if err != nil {
		logger.GetLogger().Errorf("Invalid pagination parameters: %v", err)
		httpresponse.Error(ctx, http.StatusBadRequest, "Failed to retrieve user wallets, invalid pagination parameters", err)
		return
	}

	// Get the network query parameter (if provided)
	networkStr := ctx.Query("network")
	var network *constants.NetworkType

	// Validate and parse the network type
	if networkStr != "" {
		parsedNetwork := constants.NetworkType(networkStr)
		if !constants.ValidNetworks[parsedNetwork] { // Ensure it's a valid network
			logger.GetLogger().Errorf("Invalid network parameter: %s", networkStr)
			httpresponse.Error(ctx, http.StatusBadRequest, "Invalid network parameter", nil)
			return
		}
		network = &parsedNetwork
	}

	tokenSymbols := []string{constants.USDC, constants.USDT}

	// Retrieve wallets with optional network filtering
	wallets, err := h.ucase.GetPaymentWalletsWithBalancesPagination(ctx, page, size, network, tokenSymbols)
	if err != nil {
		logger.GetLogger().Errorf("Failed to retrieve payment wallets with balances: %v", err)
		httpresponse.Error(ctx, http.StatusInternalServerError, "Failed to retrieve payment wallets with balances", err)
		return
	}

	ctx.JSON(http.StatusOK, wallets)
}

// GetReceivingWalletAddress retrieves the receiving wallet address along with native balances across networks.
// @Summary Retrieves the receiving wallet address and its native balances.
// @Description Retrieves the address of the wallet used for receiving tokens from payment wallets and its native balances across different networks.
// @Tags payment-wallet
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "Success response: {\"success\": true, \"receiving_wallet_address\": \"0x123...abc\", \"native_balances\": {\"BSC\": \"12.5\", \"AVAX C-Chain\": \"20.3\"}}"
// @Failure 500 {object} http.GeneralError "Internal server error"
// @Router /api/v1/payment-wallets/receiving-address [get]
func (h *paymentWalletHandler) GetReceivingWalletAddress(ctx *gin.Context) {
	// Retrieve the receiving wallet address and balances
	address, balances, err := h.ucase.GetReceivingWalletAddressWithBalances(
		ctx, h.config.Wallet.Mnemonic, h.config.Wallet.Passphrase, h.config.Wallet.Salt,
	)
	if err != nil {
		httpresponse.Error(ctx, http.StatusInternalServerError, "Failed to get receiving wallet address", err)
		return
	}

	// Return JSON response with both address and balances
	ctx.JSON(http.StatusOK, gin.H{
		"success":                  true,
		"receiving_wallet_address": address,
		"native_balances":          balances,
	})
}

// SyncPaymentWalletBalance syncs the balances of a specific payment wallet for multiple tokens.
// @Summary Syncs a payment wallet's balances.
// @Description Fetches the balances of a payment wallet for predefined tokens (USDT, USDC) and updates them in the database.
// @Tags payment-wallet
// @Accept json
// @Produce json
// @Param payload body dto.SyncWalletBalancePayloadDTO true "Sync wallet balance payload"
// @Success 200 {object} map[string]interface{} "Success response: {\"success\": true, \"wallet_address\": \"0x123\", \"balances\": {\"USDT\": \"100.00\", \"USDC\": \"45.00\"}}"
// @Failure 400 {object} http.GeneralError "Invalid request payload or wallet address"
// @Failure 500 {object} http.GeneralError "Internal server error"
// @Router /api/v1/payment-wallets/balance/sync [put]
func (h *paymentWalletHandler) SyncPaymentWalletBalance(ctx *gin.Context) {
	var payload dto.SyncWalletBalancePayloadDTO
	if err := ctx.ShouldBindJSON(&payload); err != nil {
		logger.GetLogger().Errorf("Invalid request payload: %v", err)
		httpresponse.Error(ctx, http.StatusBadRequest, "Invalid request payload", err)
		return
	}

	// Validate wallet address
	if !utils.IsValidEthAddress(payload.WalletAddress) {
		logger.GetLogger().Errorf("Invalid wallet address: %s", payload.WalletAddress)
		httpresponse.Error(ctx, http.StatusBadRequest, "Invalid wallet address", fmt.Errorf("invalid wallet address: %s", payload.WalletAddress))
		return
	}

	tokenSymbols := []string{constants.USDC, constants.USDT}

	// Sync balances for USDT and USDC
	balances, err := h.ucase.SyncWalletBalances(ctx, payload.WalletAddress, payload.Network, tokenSymbols)
	if err != nil {
		httpresponse.Error(ctx, http.StatusInternalServerError, "Failed to sync wallet balances", err)
		return
	}

	// Return response
	ctx.JSON(http.StatusOK, gin.H{
		"success":        true,
		"wallet_address": payload.WalletAddress,
		"balances":       balances,
	})
}
