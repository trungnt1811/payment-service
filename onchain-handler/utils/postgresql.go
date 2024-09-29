package utils

import (
	"errors"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"
)

// IsDuplicateTransactionError checks if the error is due to a unique constraint violation (e.g., duplicate transaction hash).
func IsDuplicateTransactionError(err error) bool {
	var pqErr *pgconn.PgError
	// Check if the error is a PostgreSQL error and has a unique violation code
	if errors.As(err, &pqErr) && pqErr.Code == "23505" {
		// Optionally, further verify if the constraint name is "unique_transaction_hash"
		if strings.Contains(pqErr.Message, "unique_transaction_hash") {
			return true
		}
	}
	return false
}
