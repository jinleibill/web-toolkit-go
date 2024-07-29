package egress

import (
	"context"
	"errors"
	"golang.org/x/sync/errgroup"
	"os"
	"os/signal"
	"syscall"
)

type WaiterFn func(ctx context.Context) error

type waiterCfg struct {
	signals []os.Signal
}

type Waiter interface {
	Add(fn WaiterFn)
	Wait() error
}

var _ Waiter = (*waiter)(nil)

type waiter struct {
	ctx    context.Context
	group  *errgroup.Group
	cancel context.CancelFunc
}

func NewWaiter(options ...WaiterOption) Waiter {
	ctx, cancel := context.WithCancel(context.Background())
	group, gCtx := errgroup.WithContext(ctx)

	cfg := &waiterCfg{
		signals: []os.Signal{os.Interrupt, syscall.SIGINT, syscall.SIGTERM},
	}

	for _, option := range options {
		option(cfg)
	}

	w := &waiter{
		ctx:    gCtx,
		group:  group,
		cancel: cancel,
	}

	w.group.Go(func() error {
		defer w.cancel()

		s := make(chan os.Signal, 1)
		signal.Notify(s, cfg.signals...)

		select {
		case <-s:
		case <-w.ctx.Done():
		}

		return nil
	})

	return w
}

func (w *waiter) Add(fn WaiterFn) {
	w.group.Go(func() error { return fn(w.ctx) })
}

func (w *waiter) Wait() error {
	if err := w.group.Wait(); err != nil && !errors.Is(err, context.Canceled) {
		return err
	}

	return nil
}

type WaiterOption func(*waiterCfg)

func WithSignals(signals ...os.Signal) WaiterOption {
	return func(cfg *waiterCfg) {
		cfg.signals = signals
	}
}
