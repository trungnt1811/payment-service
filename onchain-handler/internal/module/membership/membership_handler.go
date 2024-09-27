package membership

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/genefriendway/onchain-handler/internal/interfaces"
	util "github.com/genefriendway/onchain-handler/internal/utils"
	"github.com/genefriendway/onchain-handler/internal/utils/log"
)

// MembershipHandler handles membership-related requests.
type MembershipHandler struct {
	UCase interfaces.MembershipUCase
}

// NewMembershipHandler initializes a new MembershipHandler.
func NewMembershipHandler(ucase interfaces.MembershipUCase) *MembershipHandler {
	return &MembershipHandler{
		UCase: ucase,
	}
}

// GetMembershipEventsByOrderIDs retrieves membership events by a list of order IDs.
// @Summary Retrieve membership events by order IDs
// @Description This endpoint fetches a list of membership events based on the provided comma-separated list of order IDs.
// @Tags membership
// @Accept json
// @Produce json
// @Param orderIds query string true "Comma-separated list of order IDs"
// @Success 200 {array} dto.MembershipEventDTO "Successful retrieval of membership events"
// @Failure 400 {object} util.GeneralError "Invalid Order IDs or missing Order IDs"
// @Failure 500 {object} util.GeneralError "Internal server error"
// @Router /api/v1/membership/events [get]
func (h *MembershipHandler) GetMembershipEventsByOrderIDs(ctx *gin.Context) {
	// Extract order IDs from query params and split by comma.
	orderIDsStr := ctx.Query("orderIds")
	if orderIDsStr == "" {
		log.LG.Errorf("Order IDs are required")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Order IDs are required"})
		return
	}

	// Parse the comma-separated IDs into uint64 slice.
	orderIDs, err := parseOrderIDs(orderIDsStr)
	if err != nil {
		log.LG.Errorf("Invalid Order IDs: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Order IDs"})
		return
	}

	// Fetch the membership events using the use case.
	events, err := h.UCase.GetMembershipEventsByOrderIDs(ctx, orderIDs)
	if err != nil {
		log.LG.Errorf("Failed to retrieve membership events: %v", err)
		util.RespondError(ctx, http.StatusInternalServerError, "Failed to retrieve membership events", err)
		return
	}

	if len(events) == 0 {
		log.LG.Warn("No membership events found for provided Order IDs")
		ctx.JSON(http.StatusOK, []interface{}{}) // Respond with an empty array and 200 status code.
		return
	}

	// Return the event data as a JSON response.
	ctx.JSON(http.StatusOK, events)
}

// parseOrderIDs parses a comma-separated string of order IDs into a slice of uint64.
func parseOrderIDs(orderIDsStr string) ([]uint64, error) {
	var orderIDs []uint64
	idStrs := strings.Split(orderIDsStr, ",")
	for _, idStr := range idStrs {
		orderID, err := strconv.ParseUint(strings.TrimSpace(idStr), 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid Order ID: %v", err)
		}
		orderIDs = append(orderIDs, orderID)
	}
	return orderIDs, nil
}
