package usecase

import (
	"context"
	"fmt"
	"log/slog"
	"oc-data-be-challenge/internal/client"
	"oc-data-be-challenge/internal/data/dto"
	"oc-data-be-challenge/internal/data/iter"
	"oc-data-be-challenge/internal/data/repository"
	"time"
)

type DataPointUseCase struct {
	repo             *repository.DataPoint
	dataServerClient *client.DataServerClient
	logger           *slog.Logger
}

func NewDataPointUseCase(repo *repository.DataPoint, dataServerClient *client.DataServerClient) *DataPointUseCase {
	return &DataPointUseCase{repo: repo, dataServerClient: dataServerClient, logger: slog.With("component", "DataPointUseCase")}
}

func (dpuc *DataPointUseCase) Write(ctx context.Context, point dto.DataPoint) error {
	return dpuc.repo.Write(ctx, point)
}

func (dpuc *DataPointUseCase) WriteDiscard(ctx context.Context, point dto.DataPoint) error {
	return dpuc.repo.WriteDiscard(ctx, point)
}

func (dpuc *DataPointUseCase) Read(ctx context.Context) (client.DataPoint, error) {
	dp, err := dpuc.dataServerClient.DataPoint(ctx)
	if err != nil {
		return client.DataPoint{}, fmt.Errorf("failed to get datapoint from data server: %w", err)
	}

	return dp, nil
}

func (dpuc *DataPointUseCase) Collect(ctx context.Context) error {
	dp, err := dpuc.Read(ctx)
	if err != nil {
		return fmt.Errorf("failed to read datapoint: %w", err)
	}

	if dp.Time.Value.Before(time.Now().Add(-1 * time.Hour)) {
		dpuc.logger.InfoContext(ctx, "Dropping datapoint", "reason", "timestamp too old", "t", dp.Time.Value)
		return dpuc.repo.WriteDiscard(ctx, dto.DataPoint{
			Time:       dp.Time.Value,
			Value:      dp.Value.Value,
			Tags:       dp.Tags.Value,
			ReceivedAt: time.Now(),
		})
	}

	for _, value := range dp.Tags.Value {
		// drop data points with tag "system" or "suspect"
		if value == "system" || value == "suspect" {
			dpuc.logger.InfoContext(ctx, "Dropping datapoint", "tag", value, "t", dp.Time.Value)
			return dpuc.repo.WriteDiscard(ctx, dto.DataPoint{
				Time:       dp.Time.Value,
				Value:      dp.Value.Value,
				Tags:       dp.Tags.Value,
				ReceivedAt: time.Now(),
			})
		}
	}

	return dpuc.repo.Write(ctx, dto.DataPoint{
		Time:       dp.Time.Value,
		Value:      dp.Value.Value,
		Tags:       dp.Tags.Value,
		ReceivedAt: time.Now(),
	})
}

func (dpuc *DataPointUseCase) Query(ctx context.Context, start, until *time.Time) (*iter.DataPointIter, error) {
	resultIter, err := dpuc.repo.Query(ctx, start, until)
	if err != nil {
		return nil, fmt.Errorf("failed to query datapoints: %w", err)
	}

	return resultIter, nil
}
