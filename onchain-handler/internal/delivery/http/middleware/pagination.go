package middleware

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/genefriendway/onchain-handler/constants"
)

// Default Create a new pagination middleware with default values
func DefaultPagination() gin.HandlerFunc {
	return New(
		constants.DEFAULT_PAGE_TEXT,
		constants.DEFAULT_SIZE_TEXT,
		constants.DEFAULT_PAGE,
		constants.DEFAULT_PAGE_SIZE,
		constants.DEFAULT_MIN_PAGESIZE,
		constants.DEFAULT_MAX_PAGESIZE,
	)
}

// New Create a new pagniation middleware with custom values
func New(pageText, sizeText, defaultPage, defaultPageSize string, minPageSize, maxPageSize int) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract the page from the query string and convert it to an integer
		pageStr := c.DefaultQuery(pageText, defaultPage)
		page, err := strconv.Atoi(pageStr)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "page number must be an integer"})
			return
		}

		// Validate for positive page number
		if page < 0 {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "page number must be positive"})
			return
		}

		// Extract the size from the query string and convert it to an integer
		sizeStr := c.DefaultQuery(sizeText, defaultPageSize)
		size, err := strconv.Atoi(sizeStr)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "page size must be an integer"})
			return
		}

		// Validate for min and max page size
		if size < minPageSize || size > maxPageSize {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "page size must be between " + strconv.Itoa(minPageSize) + " and " + strconv.Itoa(maxPageSize)})
			return
		}

		// Set the page and size in the gin context
		c.Set(pageText, page)
		c.Set(sizeText, size)

		c.Next()
	}
}
