package closer

import (
	"time"
)

type HandlerOption func(h *Handler)

func WithTimeout(timeout time.Duration) HandlerOption {
	return func(h *Handler) {
		if timeout > 0 {
			h.Timeout = timeout
		}
	}
}

func WithForceClose(fn func()) HandlerOption {
	return func(h *Handler) {
		if fn != nil {
			h.ForceFn = fn
		}
	}
}
