package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/genefriendway/onchain-handler/internal/interfaces"
	httpresponse "github.com/genefriendway/onchain-handler/pkg/http/response"
	"github.com/genefriendway/onchain-handler/pkg/logger"
)

type metadataHandler struct {
	ucase interfaces.NetworkMetadataUCase
}

func NewMetadataHandler(ucase interfaces.NetworkMetadataUCase) *metadataHandler {
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
// @Failure 500 {object} response.GeneralError "Internal server error"
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
