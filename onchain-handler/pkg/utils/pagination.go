package utils

import (
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/genefriendway/onchain-handler/constants"
)

// ParsePaginationParams parses pagination parameters from the request context.
func ParsePaginationParams(ctx *gin.Context) (int, int, error) {
	page := ctx.DefaultQuery("page", constants.DefaultPage)
	size := ctx.DefaultQuery("size", constants.DefaultPageSize)

	pageInt, err := strconv.Atoi(page)
	if err != nil || pageInt < 1 {
		return 0, 0, fmt.Errorf("invalid page number")
	}

	sizeInt, err := strconv.Atoi(size)
	if err != nil || sizeInt < 1 {
		return 0, 0, fmt.Errorf("invalid page size")
	}

	return pageInt, sizeInt, nil
}
