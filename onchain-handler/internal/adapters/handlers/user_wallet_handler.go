package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/genefriendway/onchain-handler/conf"
	"github.com/genefriendway/onchain-handler/constants"
	"github.com/genefriendway/onchain-handler/internal/domain/dto"
	"github.com/genefriendway/onchain-handler/internal/interfaces"
	"github.com/genefriendway/onchain-handler/pkg/crypto"
	httpresponse "github.com/genefriendway/onchain-handler/pkg/http"
	"github.com/genefriendway/onchain-handler/pkg/logger"
	"github.com/genefriendway/onchain-handler/pkg/utils"
)

type userWalletHandler struct {
	ucase  interfaces.UserWalletUCase
	config *conf.Configuration
}

func NewUserWalletHandler(
	ucase interfaces.UserWalletUCase, config *conf.Configuration,
) *userWalletHandler {
	return &userWalletHandler{
		ucase:  ucase,
		config: config,
	}
}

// CreateUserWallets creates wallets for a list of user IDs.
// @Summary Create wallets for users
// @Description This endpoint allows creating wallets for a list of user IDs.
// @Tags user-wallet
// @Accept json
// @Produce json
// @Param user_ids body []string true "List of user IDs"
// @Success 201 {object} map[string]interface{} "Success created: {\"success\": true, \"data\": []dto.UserWalletPayloadDTO}"
// @Failure 400 {object} http.GeneralError "Invalid payload"
// @Failure 500 {object} http.GeneralError "Internal server error"
// @Router /api/v1/user-wallets [post]
func (h *userWalletHandler) CreateUserWallets(ctx *gin.Context) {
	var userIDs []string

	// Parse and validate the request payload (list of user IDs)
	if err := ctx.ShouldBindJSON(&userIDs); err != nil {
		logger.GetLogger().Errorf("Invalid payload: %v", err)
		httpresponse.Error(ctx, http.StatusBadRequest, "Failed to create user wallets, invalid payload:", fmt.Errorf("invalid payload: %v", err))
		return
	}

	// Ensure each user ID is a valid integer
	for _, userID := range userIDs {
		if _, err := strconv.ParseUint(userID, 10, 64); err != nil {
			logger.GetLogger().Errorf("Validation failed: User ID '%s' is not a valid integer", userID)
			httpresponse.Error(ctx, http.StatusBadRequest, "Failed to create user wallets, validation failed", fmt.Errorf("user ID '%s' is not a valid integer", userID))
			return
		}
	}

	mnemonic := h.config.Wallet.Mnemonic
	passphrase := h.config.Wallet.Passphrase
	salt := h.config.Wallet.Salt

	var payloads []dto.UserWalletPayloadDTO
	for _, userIDString := range userIDs {
		// Convert userID string to uint64
		userID, err := strconv.ParseUint(userIDString, 10, 64)
		if err != nil {
			logger.GetLogger().Errorf("Failed to parse user ID '%s' as integer: %v", userIDString, err)
			httpresponse.Error(ctx, http.StatusBadRequest, "Failed to create user wallets, validation failed", fmt.Errorf("user ID '%s' is not a valid integer", userIDString))
			return
		}

		// Generate the account based on mnemonic, passphrase, and salt
		account, _, err := crypto.GenerateAccount(mnemonic, passphrase, salt, constants.UserWallet, userID)
		if err != nil {
			logger.GetLogger().Errorf("Failed to generate account for user ID '%d': %v", userID, err)
			httpresponse.Error(ctx, http.StatusInternalServerError, "Failed to create user wallets, failed to generate account", err)
			return
		}

		// Create the payload with user ID and wallet address
		payload := dto.UserWalletPayloadDTO{
			UserID:  userID,
			Address: account.Address.Hex(),
		}
		payloads = append(payloads, payload)
	}

	// Call the use case to create wallets for the provided user IDs
	if err := h.ucase.CreateUserWallets(ctx, payloads); err != nil {
		logger.GetLogger().Errorf("Failed to create user wallets: %v", err)
		httpresponse.Error(ctx, http.StatusInternalServerError, "Failed to create user wallets", err)
		return
	}

	// Respond with success and response data
	ctx.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    payloads,
	})
}

// GetUserWallets retrieves wallets for a list of user IDs.
// @Summary Retrieve user wallets
// @Description This endpoint retrieves wallets based on a list of user IDs.
// @Tags user-wallet
// @Accept json
// @Produce json
// @Param page query int false "Page number, default is 1"
// @Param size query int false "Page size, default is 10"
// @Param user_ids query []string false "List of user IDs to filter"
// @Success 200 {object} dto.PaginationDTOResponse "Successful retrieval of user wallets"
// @Failure 400 {object} http.GeneralError "Invalid parameters"
// @Failure 500 {object} http.GeneralError "Internal server error"
// @Router /api/v1/user-wallets [get]
func (h *userWalletHandler) GetUserWallets(ctx *gin.Context) {
	// Parse pagination parameters
	page, size, err := utils.ParsePaginationParams(ctx)
	if err != nil {
		logger.GetLogger().Errorf("Invalid pagination parameters: %v", err)
		httpresponse.Error(ctx, http.StatusBadRequest, "Failed to retrieve user wallets, invalid pagination parameters", err)
		return
	}

	// Extract and parse user IDs from query string
	userIDsStr := ctx.Query("user_ids")
	var userIDs []string
	if userIDsStr != "" {
		userIDs = strings.Split(userIDsStr, ",")
	}

	// Validate each user ID to ensure it is a valid uint64
	for _, userID := range userIDs {
		if _, err := strconv.ParseUint(userID, 10, 64); err != nil {
			logger.GetLogger().Errorf("Invalid user ID: %v", userID)
			httpresponse.Error(ctx, http.StatusBadRequest, "Failed to retrieve user wallets, invalid user ID", fmt.Errorf("invalid user ID: %v", userID))
			return
		}
	}

	// Call the use case to get user wallets
	response, err := h.ucase.GetUserWallets(ctx, page, size, userIDs)
	if err != nil {
		logger.GetLogger().Errorf("Failed to retrieve user wallets: %v", err)
		httpresponse.Error(ctx, http.StatusInternalServerError, "Failed to retrieve user wallets", err)
		return
	}

	// Return the response as a JSON response
	ctx.JSON(http.StatusOK, response)
}
