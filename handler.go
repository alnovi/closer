package closer

import (
	"context"
	"errors"
	"fmt"
)

var (
	ErrHandlerTimeout   = errors.New("handler timeout")
	ErrHandlerCancelled = errors.New("handler cancelled")
	ErrHandlerPanic     = errors.New("handler panic")
)

type Handler struct {
	Name    string
	CloseFn CloseFunc
}

func NewHandler(name string, fn CloseFunc) *Handler {
	return &Handler{Name: name, CloseFn: fn}
}

func NewWrapHandler(name string, fn func()) *Handler {
	return NewHandler(name, func(_ context.Context) error {
		fn()
		return nil
	})
}

func (h *Handler) Close(ctx context.Context) error {
	var err error

	if ctx.Err() != nil {
		return ErrHandlerCancelled
	}

	done := make(chan struct{})
	go func() {
		defer func() {
			if recErr, ok := recover().(error); ok {
				err = fmt.Errorf("%w: %w", ErrHandlerPanic, recErr)
			}
			close(done)
		}()

		err = h.CloseFn(ctx)
	}()

	select {
	case <-done:
		return err
	case <-ctx.Done():
		return ErrHandlerTimeout
	}
}
