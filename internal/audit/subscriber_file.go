package audit

import (
	"encoding/json"
	"fmt"
	"os"
)

type FileSubscriber struct {
	file *os.File
	done chan struct{}
}

func NewFileSubscriber(p *Publisher, id string, path string) (*FileSubscriber, error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return nil, fmt.Errorf("audit: open file %s: %w", path, err)
	}

	s := &FileSubscriber{
		file: f,
		done: make(chan struct{}),
	}

	ch := p.Subscribe(id)
	go s.run(ch)

	return s, nil
}

func (s *FileSubscriber) run(ch <-chan Event) {
	defer close(s.done)
	defer s.file.Close()

	enc := json.NewEncoder(s.file)
	for event := range ch {
		if err := enc.Encode(event); err != nil {
			fmt.Fprintf(os.Stderr, "audit: write file error %v\n", err)
		}
	}
}

func (s *FileSubscriber) Wait() {
	<-s.done
}
