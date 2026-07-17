package audit

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"go.uber.org/zap"
)

// FileSubscriber writes audit events to a file as JSON lines.
type FileSubscriber struct {
	file   *os.File
	logger *zap.Logger
	mu     sync.Mutex
	done   chan struct{}
	cancel context.CancelFunc
}

// NewFileSubscriber subscribes to the publisher and starts writing incoming events to the given file.
func NewFileSubscriber(ctx context.Context, p *Publisher, logger *zap.Logger, id string, path string) (*FileSubscriber, error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return nil, fmt.Errorf("audit: open file %s: %w", path, err)
	}
	ctx, cancel := context.WithCancel(ctx)
	s := &FileSubscriber{
		file:   f,
		logger: logger,
		done:   make(chan struct{}),
		cancel: cancel,
	}

	ch := p.Subscribe(id)

	go s.run(ctx, ch)

	return s, nil
}

func (s *FileSubscriber) run(ctx context.Context, ch <-chan Event) {
	defer close(s.done)
	defer s.file.Close()

	enc := json.NewEncoder(s.file)

	for {
		select {
		case event, ok := <-ch:

			if !ok {
				return
			}
			s.mu.Lock()
			if err := enc.Encode(event); err != nil {
				s.logger.Error("audit: write file error", zap.Error(err))
			}
			s.mu.Unlock()
		case <-ctx.Done():
			return
		}

	}
}

// Stop cancels event processing and waits for the worker to finish.
func (s *FileSubscriber) Stop() {
	s.cancel()
	<-s.done
}

// Wait blocks until the subscriber worker finishes.
func (s *FileSubscriber) Wait() {
	<-s.done
}
