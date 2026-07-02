package audit

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/go-resty/resty/v2"
)

type URLSubscriber struct {
	client *resty.Client
	done   chan struct{}
	cancel context.CancelFunc
}

func NewURLSubscriber(ctx context.Context, p *Publisher, id string, url string) (*URLSubscriber, error) {
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
		SetContentLength(true)
	ctx, cancel := context.WithCancel(ctx)
	s := &URLSubscriber{
		client: client,
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
			fmt.Fprintf(os.Stderr, "audit: url send error: %v\n", err)
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

func (s *URLSubscriber) Stop() {
	s.cancel()
	<-s.done
}

func (s *URLSubscriber) Wait() {
	<-s.done
}
