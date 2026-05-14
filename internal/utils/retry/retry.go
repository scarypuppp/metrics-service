package retry

import (
	"time"
)

var intervals = []time.Duration{
	1 * time.Second,
	3 * time.Second,
	5 * time.Second,
}

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
	return fn()
}
