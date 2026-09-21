package closer

import (
	"context"
	"errors"
	"fmt"
	"time"
)

var (
	ErrHandlerTimeout   = errors.New("handler timeout")
	ErrHandlerCancelled = errors.New("handler cancelled")
	ErrHandlerPanic     = errors.New("handler panic")
)

type Handler struct {
	Name    string
	Timeout time.Duration
	CloseFn func(ctx context.Context) error
	ForceFn func()
}

func NewCtxHandlerErr(name string, fn func(ctx context.Context) error, opts ...HandlerOption) *Handler {
	h := &Handler{Name: name, Timeout: time.Minute, CloseFn: fn}
	for _, opt := range opts {
		if opt != nil {
			opt(h)
		}
	}
	return h
}

func NewCtxHandler(name string, fn func(ctx context.Context), opts ...HandlerOption) *Handler {
	closeFn := func(ctx context.Context) error {
		fn(ctx)
		return nil
	}
	return NewCtxHandlerErr(name, closeFn, opts...)
}

func NewHandlerErr(name string, fn func() error, opts ...HandlerOption) *Handler {
	closeFn := func(ctx context.Context) error {
		return fn()
	}
	return NewCtxHandlerErr(name, closeFn, opts...)
}

func NewHandler(name string, fn func(), opts ...HandlerOption) *Handler {
	closeFn := func(ctx context.Context) error {
		fn()
		return nil
	}
	return NewCtxHandlerErr(name, closeFn, opts...)
}

func (h *Handler) Close(ctx context.Context) error {
	var err error

	if ctx.Err() != nil {
		return ErrHandlerCancelled
	}

	ctx, cancel := context.WithTimeout(ctx, h.Timeout)
	defer cancel()

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
		if h.ForceFn != nil {
			go func() {
				defer func() {
					_ = recover()
				}()
				h.ForceFn()
			}()
		}
		return ErrHandlerTimeout
	}
}
