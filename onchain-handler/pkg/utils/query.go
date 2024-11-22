package utils

import (
	"fmt"
	"strings"

	"github.com/genefriendway/onchain-handler/constants"
)

func ParseOptionalQuery(param string) *string {
	if param == "" {
		return nil
	}
	return &param
}

func ParseSortParameter(sort string) (*string, constants.OrderDirection, error) {
	if sort == "" {
		// Default sorting
		defaultField := "id"
		return &defaultField, constants.Asc, nil
	}

	// Find the last underscore in the input, which separates field and direction
	lastUnderscore := strings.LastIndex(sort, "_")
	if lastUnderscore == -1 {
		return nil, "", fmt.Errorf("invalid format, expected field_direction (e.g., id_asc)")
	}

	// Extract the field name and direction
	orderBy := sort[:lastUnderscore]                           // Field is everything before the last underscore
	orderDirection := strings.ToUpper(sort[lastUnderscore+1:]) // Direction is everything after the last underscore

	// Validate the direction
	switch orderDirection {
	case "ASC":
		return &orderBy, constants.Asc, nil
	case "DESC":
		return &orderBy, constants.Desc, nil
	default:
		return nil, "", fmt.Errorf("invalid direction, expected ASC or DESC")
	}
}
