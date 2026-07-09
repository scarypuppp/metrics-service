package retry

import (
	"errors"
	"time"
)

// ErrMaxRetriesExceeded is returned by Do when all retry attempts fail with retriable errors.
var ErrMaxRetriesExceeded = errors.New("max retries exceeded")

var intervals = []time.Duration{
	1 * time.Second,
	3 * time.Second,
	5 * time.Second,
}

// Do runs fn, retrying with increasing backoff intervals as long as isRetriable reports the error as retriable.
func Do(fn func() error, isRetriable func(error) bool) error {
	if err := fn(); err == nil {
		return nil
	} else if !isRetriable(err) {
		return err
	}
	for _, interval := range intervals {
		time.Sleep(interval)
		if err := fn(); err == nil {
			return nil
		} else if !isRetriable(err) {
			return err
		}
	}
	return ErrMaxRetriesExceeded
}
