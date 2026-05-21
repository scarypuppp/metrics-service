package retry

import (
	"errors"
	"net"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
)

func IsNetworkError(err error) bool {
	if err == nil {
		return false
	}
	var netErr net.Error
	return errors.As(err, &netErr)
}

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
