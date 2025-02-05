package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/genefriendway/onchain-handler/conf"
	"github.com/genefriendway/onchain-handler/constants"
	"github.com/genefriendway/onchain-handler/internal/dto"
	"github.com/genefriendway/onchain-handler/internal/interfaces"
	httpresponse "github.com/genefriendway/onchain-handler/pkg/http/response"
	"github.com/genefriendway/onchain-handler/pkg/logger"
	"github.com/genefriendway/onchain-handler/pkg/utils"
)

type paymentWalletHandler struct {
	ucase  interfaces.PaymentWalletUCase
	config *conf.Configuration
}

func NewPaymentWalletHandler(ucase interfaces.PaymentWalletUCase, config *conf.Configuration) *paymentWalletHandler {
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
// @Param network query string false "Filter by network (e.g., BSC, AVAX C-Chain)"
// @Success 200 {object} dto.PaymentWalletBalanceDTO
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

	wallet, err := h.ucase.GetPaymentWalletByAddress(ctx, network, address)
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
// @Failure 400 {object} response.GeneralError "Invalid network"
// @Failure 500 {object} response.GeneralError "Internal server error"
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

	// Retrieve wallets with optional network filtering
	nonZeroOnly := true
	wallets, err := h.ucase.GetPaymentWalletsWithBalancesPagination(ctx, page, size, nonZeroOnly, network)
	if err != nil {
		logger.GetLogger().Errorf("Failed to retrieve payment wallets with balances: %v", err)
		httpresponse.Error(ctx, http.StatusInternalServerError, "Failed to retrieve payment wallets with balances", err)
		return
	}

	ctx.JSON(http.StatusOK, wallets)
}

// GetReceivingWalletAddress retrieves the receiving wallet address.
// @Summary Retrieves the receiving wallet address.
// @Description Retrieves the address of the wallet used for receiving tokens from payment wallets.
// @Tags payment-wallet
// @Accept json
// @Produce json
// @Success 200 {string} string "Receiving wallet address"
// @Failure 500 {object} response.GeneralError "Internal server error"
// @Router /api/v1/payment-wallets/receiving-address [get]
func (h *paymentWalletHandler) GetReceivingWalletAddress(ctx *gin.Context) {
	address, err := h.ucase.GetReceivingWalletAddress(
		ctx, h.config.Wallet.Mnemonic, h.config.Wallet.Passphrase, h.config.Wallet.Salt,
	)
	if err != nil {
		httpresponse.Error(ctx, http.StatusInternalServerError, "Failed to get receiving wallet address", err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"receiving_wallet_address": address})
}

// SyncPaymentWalletBalance syncs the balance of a specific payment wallet.
// @Summary Syncs a payment wallet balance.
// @Description Fetches the balance of a payment wallet for a specific token and updates it in the database.
// @Tags payment-wallet
// @Accept json
// @Produce json
// @Param payload body dto.SyncWalletBalancePayload true "Sync wallet balance payload"
// @Success 200 {object} map[string]interface{} "Success response: {\"success\": true, \"wallet_address\": \"0x123\", \"usdt_amount\": 100.00}"
// @Failure 400 {object} response.GeneralError "Invalid request payload or token symbol"
// @Failure 500 {object} response.GeneralError "Internal server error"
// @Router /api/v1/payment-wallets/balance/sync [put]
func (h *paymentWalletHandler) SyncPaymentWalletBalance(ctx *gin.Context) {
	var payload dto.SyncWalletBalancePayload
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

	// Call the use case to fetch the wallet balance
	usdtAmount, err := h.ucase.SyncWalletBalance(ctx, payload.WalletAddress, payload.Network)
	if err != nil {
		httpresponse.Error(ctx, http.StatusInternalServerError, "Failed to sync wallet balance", err)
		return
	}

	// Return the balance as a JSON response
	ctx.JSON(http.StatusOK, gin.H{
		"success":        true,
		"wallet_address": payload.WalletAddress,
		"usdt_amount":    usdtAmount,
	})
}
