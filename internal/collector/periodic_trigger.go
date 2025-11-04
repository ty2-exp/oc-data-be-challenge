package collector

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

type PeriodicTrigger struct {
	interval             time.Duration
	stopCh               chan struct{}
	startOnce            *sync.Once
	stopOnce             *sync.Once
	triggerCtx           context.Context
	triggerCtxCancelFunc context.CancelFunc
	triggerFn            func(ctx context.Context) error
	logger               *slog.Logger
}

func NewPeriodicTrigger(name string, triggerFn func(ctx context.Context) error, interval time.Duration) *PeriodicTrigger {
	periodicTrigger := &PeriodicTrigger{
		triggerFn: triggerFn,
		interval:  interval,
		logger:    slog.With("component", "PeriodicTrigger", "name", name),
	}

	periodicTrigger.reset()
	return periodicTrigger
}

func (pt *PeriodicTrigger) reset() {
	pt.stopCh = make(chan struct{}, 1)
	pt.startOnce = &sync.Once{}
	pt.stopOnce = &sync.Once{}
	pt.triggerCtx, pt.triggerCtxCancelFunc = context.WithCancel(context.Background())
}

func (pt *PeriodicTrigger) Start() {
	pt.startOnce.Do(pt.start)
}

func (pt *PeriodicTrigger) start() {
	pt.logger.InfoContext(pt.triggerCtx, "PeriodicTrigger started", "interval", pt.interval)
	err := pt.triggerFn(pt.triggerCtx)
	if err != nil {
		pt.logger.ErrorContext(pt.triggerCtx, "PeriodicTrigger initial collection error", "error", err)
	}

	ticker := time.NewTicker(pt.interval)
	doneCh := make(chan struct{})

	// Goroutine to listen for stop signal
	go func() {
		select {
		case <-pt.stopCh:
			pt.logger.Info("PeriodicTrigger stopping")
			pt.triggerCtxCancelFunc()
			ticker.Stop()
			<-doneCh
		}
	}()

	for {
		select {
		case <-ticker.C:
			pt.logger.Debug("PeriodicTrigger tick")
			err := pt.triggerFn(pt.triggerCtx)
			if err != nil {
				pt.logger.Error("PeriodicTrigger collection error", "error", err)
				continue
			}
		case <-doneCh:
			break
		}
	}
}

func (pt *PeriodicTrigger) Stop() {
	pt.stopOnce.Do(pt.stop)
}

func (pt *PeriodicTrigger) stop() {
	pt.stopCh <- struct{}{}
	close(pt.stopCh)
	pt.reset()
}
