package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"

	httpresponse "github.com/genefriendway/onchain-handler/pkg/http/response"
	"github.com/genefriendway/onchain-handler/pkg/logger"
)

func ValidateVendorID() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		vendorID := ctx.GetHeader("Vendor-Id")

		if vendorID == "" {
			// Log the specific error internally
			logger.GetLogger().Info("Validation failed: Vendor-Id header is missing")
			// Return a generalized error message to the client
			httpresponse.Error(ctx, http.StatusBadRequest, "Invalid request headers", nil)
			ctx.Abort()
			return
		}

		if len(vendorID) > 33 {
			// Log the specific error internally
			logger.GetLogger().Infof("Validation failed: Vendor-Id header exceeds max length (%d characters)", len(vendorID))
			// Return a generalized error message to the client
			httpresponse.Error(ctx, http.StatusBadRequest, "Invalid request headers", nil)
			ctx.Abort()
			return
		}

		ctx.Next() // Continue to the next handler
	}
}
