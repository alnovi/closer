package closer

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"testing/synctest"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCloser(t *testing.T) {
	testCases := []struct {
		name      string
		handlers  []*Handler
		sleep     time.Duration
		expErr    string
		expReport []string
	}{
		{
			name: "Success all tasks",
			handlers: []*Handler{
				NewHandler("redis", func() { time.Sleep(10 * time.Second) }),
				NewHandler("postgres", func() { time.Sleep(30 * time.Second) }),
			},
			sleep:     time.Minute,
			expErr:    "",
			expReport: []string{"postgres", "redis"},
		},
		{
			name: "Success close with errors",
			handlers: []*Handler{
				NewCtxHandlerErr("redis", func(_ context.Context) error {
					time.Sleep(10 * time.Second)
					return nil
				}),
				NewCtxHandlerErr("postgres", func(_ context.Context) error {
					time.Sleep(30 * time.Second)
					return errors.New("some error")
				}),
			},
			sleep:     2 * time.Minute,
			expErr:    ErrShutdownWithErrors.Error(),
			expReport: []string{"postgres - some error", "redis"},
		},
		{
			name: "Redis is timeout",
			handlers: []*Handler{
				NewCtxHandler("redis", func(_ context.Context) { time.Sleep(20 * time.Second) }),
				NewCtxHandler("postgres", func(_ context.Context) { time.Sleep(20 * time.Second) }),
				NewCtxHandler("http", func(_ context.Context) { time.Sleep(30 * time.Second) }),
			},
			sleep:     2 * time.Minute,
			expErr:    ErrShutdownCancelled.Error(),
			expReport: []string{"http", "postgres", "redis - handler timeout"},
		},
		{
			name: "Redis is cancelled",
			handlers: []*Handler{
				NewHandlerErr("redis", func() error { time.Sleep(10 * time.Second); return nil }),
				NewHandlerErr("postgres", func() error { time.Sleep(20 * time.Second); return nil }),
				NewHandlerErr("http", func() error { time.Sleep(20 * time.Second); return nil }),
				NewHandlerErr("grpc", func() error { time.Sleep(30 * time.Second); return nil }),
			},
			sleep:     2 * time.Minute,
			expErr:    ErrShutdownCancelled.Error(),
			expReport: []string{"grpc", "http", "postgres - handler timeout", "redis - handler cancelled"},
		},
		{
			name: "Success with panic",
			handlers: []*Handler{
				NewHandler("redis", func() {
					time.Sleep(10 * time.Second)
					panic(errors.New("some panic"))
				}),
			},
			sleep:     time.Minute,
			expErr:    ErrShutdownWithErrors.Error(),
			expReport: []string{"redis - handler panic: some panic"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				closer := New(0)
				for _, handler := range tc.handlers {
					closer.Add(handler)
				}

				var report *Report
				var err error

				require.NotPanics(t, func() {
					go func() {
						report, err = closer.Close()
					}()
				}, "handler close panic")

				time.Sleep(tc.sleep)
				synctest.Wait()

				if err != nil || tc.expErr != "" {
					require.Error(t, err, "closer was expected to error")
					require.NotEmptyf(t, tc.expErr, "closer expected not zero error")
					require.ErrorContains(t, err, tc.expErr, "closer not contains error")
				}

				var actReport []string

				for _, item := range report.Items {
					if item.Error != nil {
						actReport = append(actReport, fmt.Sprintf("%s - %s", item.Name, item.Error))
					} else {
						actReport = append(actReport, item.Name)
					}
				}

				assert.Equal(t, tc.expReport, actReport)
			})
		})
	}
}
