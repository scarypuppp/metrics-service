package retry

import (
	"errors"
	"net"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
)

// IsNetworkError reports whether err is a network error suitable for retrying.
func IsNetworkError(err error) bool {
	if err == nil {
		return false
	}
	var netErr net.Error
	return errors.As(err, &netErr)
}

// IsPgConnectionError reports whether err is a PostgreSQL connection error suitable for retrying.
func IsPgConnectionError(err error) bool {
	if err == nil {
		return false
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgerrcode.IsConnectionException(pgErr.Code)
	}
	return false
}
