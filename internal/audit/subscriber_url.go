package audit

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-resty/resty/v2"
	"go.uber.org/zap"
)

// URLSubscriber sends audit events as JSON to a remote HTTP endpoint.
type URLSubscriber struct {
	client *resty.Client
	logger *zap.Logger
	done   chan struct{}
	cancel context.CancelFunc
}

// NewURLSubscriber subscribes to the publisher and starts posting incoming events to the given URL.
func NewURLSubscriber(ctx context.Context, p *Publisher, logger *zap.Logger, id string, url string) (*URLSubscriber, error) {
	client := resty.NewWithClient(&http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 10,
			IdleConnTimeout:     90 * time.Second,
		},
	})
	client.SetBaseURL(url).
		SetHeader("Content-Type", "application/json").
		SetContentLength(true).
		SetRetryCount(3).
		SetRetryWaitTime(1 * time.Second).
		SetRetryMaxWaitTime(5 * time.Second)

	client.AddRetryCondition(
		func(r *resty.Response, err error) bool {
			return err != nil || r.StatusCode() >= 500
		},
	)

	ctx, cancel := context.WithCancel(ctx)
	s := &URLSubscriber{
		client: client,
		logger: logger,
		cancel: cancel,
		done:   make(chan struct{}),
	}

	ch := p.Subscribe(id)
	go s.run(ctx, ch)
	return s, nil
}

func (s *URLSubscriber) run(ctx context.Context, ch <-chan Event) {
	defer close(s.done)

	for event := range ch {
		if err := s.send(ctx, event); err != nil {
			if errors.Is(err, context.Canceled) {
				return
			}
			s.logger.Error("audit: url send error", zap.Error(err))
		}
	}
}

func (s *URLSubscriber) send(ctx context.Context, event Event) error {
	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("audit: marshal event: %w", err)
	}
	req := s.client.R().SetContext(ctx).SetBody(body)
	resp, err := req.Post("")
	if err != nil {
		return fmt.Errorf("audit: failed to send metric: %w", err)
	}
	if resp.IsError() {
		return fmt.Errorf("audit: server returned %d: %s", resp.StatusCode(), resp.String())
	}
	return nil
}

// Stop cancels event processing and waits for the worker to finish.
func (s *URLSubscriber) Stop() {
	s.cancel()
	<-s.done
}

// Wait blocks until the subscriber worker finishes.
func (s *URLSubscriber) Wait() {
	<-s.done
}
