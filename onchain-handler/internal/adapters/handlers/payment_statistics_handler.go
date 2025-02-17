package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/genefriendway/onchain-handler/constants"
	"github.com/genefriendway/onchain-handler/internal/interfaces"
	httpresponse "github.com/genefriendway/onchain-handler/pkg/http"
	"github.com/genefriendway/onchain-handler/pkg/logger"
	"github.com/genefriendway/onchain-handler/pkg/utils"
)

type paymentStatisticsHandler struct {
	ucase interfaces.PaymentStatisticsUCase
}

func NewPaymentStatisticsHandler(ucase interfaces.PaymentStatisticsUCase) *paymentStatisticsHandler {
	return &paymentStatisticsHandler{
		ucase: ucase,
	}
}

// GetPaymentStatistics retrieves payment statistics by granularity and time range.
// @Summary Retrieve payment statistics
// @Description This endpoint retrieves payment statistics based on granularity and time range.
// @Tags payment-statistics
// @Accept json
// @Produce json
// @Param Vendor-Id header string true "Vendor ID for authentication"
// @Param granularity query string true "Granularity (DAILY, WEEKLY, MONTHLY, YEARLY)"
// @Param start_time query int true "Start time in UNIX timestamp format"
// @Param end_time query int true "End time in UNIX timestamp format"
// @Success 200 {object} []dto.PaymentStatistics "Payment statistics retrieved successfully"
// @Failure 400 {object} http.GeneralError "Invalid parameters"
// @Failure 500 {object} http.GeneralError "Internal server error"
// @Router /api/v1/payment-statistics [get]
func (h *paymentStatisticsHandler) GetPaymentStatistics(ctx *gin.Context) {
	// Get the Vendor-Id from the header
	vendorID := ctx.GetHeader("Vendor-Id")

	// Parse and validate granularity parameter
	granularity := ctx.Query("granularity")
	switch granularity {
	case constants.Daily, constants.Weekly, constants.Monthly, constants.Yearly:
		// Valid granularities, proceed
	default:
		httpresponse.Error(ctx, http.StatusBadRequest, "Invalid granularity. Valid options are: DAILY, WEEKLY, MONTHLY, YEARLY", nil)
		return
	}

	// Parse and validate start_time parameter
	startTime, err := utils.ParseOptionalUnixTimestamp(ctx.Query("start_time"))
	if err != nil {
		logger.GetLogger().Errorf("Invalid start_time: %v", err)
		httpresponse.Error(ctx, http.StatusBadRequest, "Invalid start_time. Provide a valid UNIX timestamp.", err)
		return
	}

	// Parse and validate end_time parameter
	endTime, err := utils.ParseOptionalUnixTimestamp(ctx.Query("end_time"))
	if err != nil {
		logger.GetLogger().Errorf("Invalid end_time: %v", err)
		httpresponse.Error(ctx, http.StatusBadRequest, "Invalid end_time. Provide a valid UNIX timestamp.", err)
		return
	}

	// Return an error if start_time is later than end_time
	if startTime.After(*endTime) {
		httpresponse.Error(ctx, http.StatusBadRequest, "start_time must be earlier than end_time", nil)
		return
	}

	// Call the use case to retrieve statistics
	paymentStatistics, err := h.ucase.GetStatisticsByTimeRangeAndGranularity(ctx, granularity, *startTime, *endTime, vendorID)
	if err != nil {
		logger.GetLogger().Errorf("Failed to retrieve payment statistics: %v", err)
		httpresponse.Error(ctx, http.StatusInternalServerError, "Failed to retrieve payment statistics", err)
		return
	}

	// Return the response
	ctx.JSON(http.StatusOK, paymentStatistics)
}
