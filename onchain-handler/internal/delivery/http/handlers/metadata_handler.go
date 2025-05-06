package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	ucasetypes "github.com/genefriendway/onchain-handler/internal/domain/ucases/types"
	httpresponse "github.com/genefriendway/onchain-handler/pkg/http"
	"github.com/genefriendway/onchain-handler/pkg/logger"
)

type metadataHandler struct {
	ucase ucasetypes.MetadataUCase
}

func NewMetadataHandler(ucase ucasetypes.MetadataUCase) *metadataHandler {
	return &metadataHandler{
		ucase: ucase,
	}
}

// GetNetworksMetadata retrieves all networks metadata.
// @Summary Retrieves all networks metadata.
// @Description Retrieves all networks metadata.
// @Tags metadata
// @Accept json
// @Produce json
// @Success 200 {array} dto.NetworkMetadataDTO
// @Failure 500 {object} http.GeneralError "Internal server error"
// @Router /api/v1/metadata/networks [get]
func (h *metadataHandler) GetNetworksMetadata(ctx *gin.Context) {
	metadata, err := h.ucase.GetNetworksMetadata(ctx)
	if err != nil {
		logger.GetLogger().Errorf("Failed to retrieve networks metadata: %v", err)
		httpresponse.Error(ctx, http.StatusBadRequest, "Failed to retrieve networks metadata", err)
		return
	}
	ctx.JSON(http.StatusOK, metadata)
}

// GetTokensMetadata retrieves all tokens metadata.
// @Summary Retrieves all tokens metadata.
// @Description Retrieves all tokens metadata.
// @Tags metadata
// @Accept json
// @Produce json
// @Success 200 {array} dto.TokenMetadataDTO
// @Failure 500 {object} http.GeneralError "Internal server error"
// @Router /api/v1/metadata/tokens [get]
func (h *metadataHandler) GetTokensMetadata(ctx *gin.Context) {
	metadata, err := h.ucase.GetTokensMetadata(ctx)
	if err != nil {
		logger.GetLogger().Errorf("Failed to retrieve tokens metadata: %v", err)
		httpresponse.Error(ctx, http.StatusBadRequest, "Failed to retrieve tokens metadata", err)
		return
	}
	ctx.JSON(http.StatusOK, metadata)
}
