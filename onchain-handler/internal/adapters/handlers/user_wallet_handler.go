package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/genefriendway/onchain-handler/conf"
	"github.com/genefriendway/onchain-handler/constants"
	"github.com/genefriendway/onchain-handler/internal/dto"
	"github.com/genefriendway/onchain-handler/internal/interfaces"
	"github.com/genefriendway/onchain-handler/internal/middleware"
	"github.com/genefriendway/onchain-handler/pkg/crypto"
	"github.com/genefriendway/onchain-handler/pkg/logger"
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
// @Success 200 {object} map[string]interface{} "Success response: {\"success\": true, \"data\": []interface{}}"
// @Failure 400 {object} map[string]interface{} "Invalid payload"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/user-wallets [post]
func (h *userWalletHandler) CreateUserWallets(ctx *gin.Context) {
	var userIDs []string

	// Parse and validate the request payload (list of user IDs)
	if err := ctx.ShouldBindJSON(&userIDs); err != nil {
		logger.GetLogger().Errorf("Invalid payload: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid payload",
			"details": err.Error(),
		})
		return
	}

	// Ensure each user ID is a valid integer
	for _, userID := range userIDs {
		if _, err := strconv.ParseUint(userID, 10, 64); err != nil {
			logger.GetLogger().Errorf("Validation failed: User ID '%s' is not a valid integer", userID)
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error":   "Validation error",
				"details": "User IDs must be valid integers",
			})
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
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error":   "Invalid user ID",
				"details": "User ID must be a valid integer",
			})
			return
		}

		// Generate the account based on mnemonic, passphrase, and salt
		account, _, err := crypto.GenerateAccount(mnemonic, passphrase, salt, constants.UserWallet, userID)
		if err != nil {
			logger.GetLogger().Errorf("Failed to generate account for user ID '%d': %v", userID, err)
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Account generation failed",
				"details": err.Error(),
			})
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
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to create user wallets",
			"details": err.Error(),
		})
		return
	}

	// Respond with success and response data
	ctx.JSON(http.StatusOK, gin.H{
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
// @Failure 400 {object} response.GeneralError "Invalid parameters"
// @Failure 500 {object} response.GeneralError "Internal server error"
// @Router /api/v1/user-wallets [get]
func (h *userWalletHandler) GetUserWallets(ctx *gin.Context) {
	// Parse pagination parameters
	page, size, err := middleware.ParsePaginationParams(ctx)
	if err != nil {
		logger.GetLogger().Errorf("Invalid pagination parameters: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error":   "Invalid user ID",
				"details": fmt.Sprintf("User ID '%s' is not a valid uint64", userID),
			})
			return
		}
	}

	// Call the use case to get user wallets
	response, err := h.ucase.GetUserWallets(ctx, page, size, userIDs)
	if err != nil {
		logger.GetLogger().Errorf("Failed to retrieve user wallets: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to retrieve user wallets",
			"details": err.Error(),
		})
		return
	}

	// Return the response as a JSON response
	ctx.JSON(http.StatusOK, response)
}
