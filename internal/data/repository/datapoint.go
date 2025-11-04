package repository

import (
	"context"
	"errors"
	"oc-data-be-challenge/internal/data/dto"
	"oc-data-be-challenge/internal/data/iter"
	"time"

	"github.com/InfluxCommunity/influxdb3-go/v2/influxdb3"
)

type DataPoint struct {
	client *influxdb3.Client
}

func NewDataPoint(client *influxdb3.Client) *DataPoint {
	return &DataPoint{client: client}
}

func (dp *DataPoint) write(ctx context.Context, point dto.DataPoint, table string) error {
	err := dp.client.WritePoints(ctx, []*influxdb3.Point{
		influxdb3.NewPoint(table,
			nil,
			map[string]any{
				"value":       point.Value,
				"tags":        point.Tags,
				"received_at": point.ReceivedAt,
			},
			point.Time,
		),
	})

	if err != nil {
		return errors.Join(errors.New("failed to write datapoint"), err)
	}
	return nil
}

func (dp *DataPoint) Write(ctx context.Context, point dto.DataPoint) error {
	return dp.write(ctx, point, "datapoint")
}

func (dp *DataPoint) WriteDiscard(ctx context.Context, point dto.DataPoint) error {
	return dp.write(ctx, point, "datapoint_discarded")
}

func (dp *DataPoint) Query(ctx context.Context, start, until *time.Time) (*iter.DataPointIter, error) {
	query := `SELECT * FROM datapoint`
	parameters := influxdb3.QueryParameters{}
	if start != nil {
		query += ` WHERE time >= $start`
		parameters["start"] = start
	}

	if until != nil {
		query += ` AND time <= $until`
		parameters["until"] = until
	}

	query += ` ORDER BY time DESC`

	resultIter, err := dp.client.QueryWithParameters(ctx, query, parameters)
	if err != nil {
		return nil, errors.Join(errors.New("failed to execute query"), err)
	}

	return iter.NewDataPointIter(resultIter), nil
}
