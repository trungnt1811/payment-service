package utils

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/genefriendway/onchain-handler/constants"
)

// ParseOptionalUnixTimestamp parses a time string in UNIX timestamp format and returns a *time.Time.
// If the input string is empty, it returns nil and no error.
func ParseOptionalUnixTimestamp(timestampStr string) (*time.Time, error) {
	if timestampStr == "" {
		// Return nil if the input is empty
		return nil, nil
	}

	// Convert the string to an integer
	timestampInt, err := strconv.ParseInt(timestampStr, 10, 64)
	if err != nil {
		// Return an error if the conversion fails
		return nil, fmt.Errorf("invalid UNIX timestamp: %w", err)
	}

	// Convert the integer to a time.Time object
	parsedTime := time.Unix(timestampInt, 0).UTC()

	// Return the parsed time as a pointer
	return &parsedTime, nil
}

// ParseOptionalTime parses a time string in RFC3339 format and returns a *time.Time.
// If the input string is empty, it returns nil and no error.
func ParseOptionalTime(timeStr string) (*time.Time, error) {
	if timeStr == "" {
		// Return nil if the input is empty
		return nil, nil
	}

	// Parse the time string using RFC3339 format
	parsedTime, err := time.Parse(time.RFC3339, timeStr)
	if err != nil {
		// Return an error if the time cannot be parsed
		return nil, err
	}

	// Return the parsed time as a pointer
	return &parsedTime, nil
}

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
