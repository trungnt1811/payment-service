package lock

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/genefriendway/onchain-handler/internal/interfaces"
	util "github.com/genefriendway/onchain-handler/internal/utils"
	"github.com/genefriendway/onchain-handler/internal/utils/log"
)

type LockHandler struct {
	LockUCase interfaces.LockUCase
}

func NewLockHandler(
	ucase interfaces.LockUCase,
) *LockHandler {
	return &LockHandler{
		LockUCase: ucase,
	}
}

// GetLatestLockEventsByUserAddress retrieves the latest lock events by a user's address.
// @Summary Get list of latest lock events by user address
// @Description This endpoint fetches a paginated list of lock events based on the provided user address.
// @Tags lock
// @Accept json
// @Produce json
// @Param userAddress query string true "User's wallet address"
// @Param page query int false "Page number, default is 1"
// @Param size query int false "Page size, default is 10"
// @Success 200 {object} dto.LockEventDTOResponse "Successful retrieval of lock events"
// @Failure 400 {object} util.GeneralError "Invalid user address or missing parameters"
// @Failure 500 {object} util.GeneralError "Internal server error"
// @Security ApiKeyAuth
// @Router /api/v1/lock/latest-events [get]
func (handler *LockHandler) GetLatestLockEventsByUserAddress(ctx *gin.Context) {
	// Extract user address from query params
	userAddress := ctx.Query("userAddress")
	if userAddress == "" {
		log.LG.Errorf("User address is required")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "User address is required"})
		return
	}

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

	// Fetch the lock events using the use case
	response, err := handler.LockUCase.GetLatestLockEventsByUserAddress(ctx, userAddress, pageInt, sizeInt)
	if err != nil {
		log.LG.Errorf("Failed to retrieve latest lock events: %v", err)
		util.RespondError(ctx, http.StatusInternalServerError, "Failed to retrieve latest lock events", err)
		return
	}

	// Return the response as a JSON response
	ctx.JSON(http.StatusOK, response)
}
