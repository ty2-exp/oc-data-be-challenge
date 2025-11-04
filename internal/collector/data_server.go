package collector

import (
	"context"
	"fmt"
	"oc-data-be-challenge/internal/usecase"
	"time"
)

func NewDataServerCollector(datapointUseCase *usecase.DataPointUseCase, interval time.Duration) *PeriodicTrigger {
	return NewPeriodicTrigger(
		"DataServerCollector",
		func(ctx context.Context) error {
			err := datapointUseCase.Collect(ctx)
			if err != nil {
				return fmt.Errorf("failed to collect data point: %w", err)
			}
			return nil
		},
		interval,
	)
}
