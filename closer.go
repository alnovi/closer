package closer

import (
	"context"
	"errors"
	"slices"
	"sync"
	"time"
)

var (
	ErrShutdownCancelled  = errors.New("shutdown cancelled")
	ErrShutdownWithErrors = errors.New("shutdown finished with errors")
)

type CloseFunc func(ctx context.Context) error

type Closer struct {
	handlers []*Handler
	timeout  time.Duration
	mu       sync.Mutex
}

func New(timeout time.Duration) *Closer {
	if timeout <= 0 {
		timeout = time.Minute
	}
	return &Closer{timeout: timeout}
}

func (c *Closer) Add(handler *Handler) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if handler != nil {
		c.handlers = append(c.handlers, handler)
	}
}

func (c *Closer) Close() (*Report, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	slices.Reverse(c.handlers)
	defer func() {
		c.handlers = nil
	}()

	report := NewReport()

	complete := make(chan struct{})
	go func() {
		for _, handler := range c.handlers {
			start := time.Now()
			err := handler.Close(ctx)
			report.AddItem(handler.Name, time.Since(start), err)
		}
		close(complete)
	}()

	select {
	case <-complete:
		break
	case <-ctx.Done():
		return report, ErrShutdownCancelled
	}

	if report.HasErrors() {
		return report, ErrShutdownWithErrors
	}

	return report, nil
}
